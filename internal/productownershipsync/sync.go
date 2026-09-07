// Package productownershipsync derives one product's ownership policy from the
// trusted product catalog and the repository's Terraform examples and docs.
package productownershipsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

const PolicyRelativePath = ".github/product-owners.json"

// Result describes one deterministic product policy update.
type Result struct {
	ProductName    string
	MasterData     productcatalog.ProductMasterDataIdentity
	Previous       productownership.Product
	HadPrevious    bool
	Generated      productownership.Product
	AddedPaths     []string
	RemovedPaths   []string
	UsersChanged   bool
	Changed        bool
	Baseline       []byte
	PolicyContents []byte
}

type policyDocument struct {
	Version  int                        `json:"version"`
	Core     json.RawMessage            `json:"core"`
	Products map[string]json.RawMessage `json:"products"`
}

type catalogIndex struct {
	metadataByName map[string]productcatalog.OwnershipMetadata
	ownerByType    map[string]string
}

type docAsset struct {
	path          string
	section       string
	stem          string
	resolvedOwner string
}

// Generate derives the target product entry without changing the repository.
func Generate(root, productName string, githubUsers []string) (Result, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Result{}, fmt.Errorf("resolve repository root: %w", err)
	}
	current, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(PolicyRelativePath)))
	if err != nil {
		return Result{}, fmt.Errorf("read product ownership policy: %w", err)
	}
	return GenerateFromPolicy(root, current, productName, githubUsers)
}

// GenerateFromPolicy derives a target product entry from an explicit baseline.
// It is suitable for trusted CI that must not execute pull request code.
func GenerateFromPolicy(root string, current []byte, productName string, githubUsers []string) (Result, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Result{}, fmt.Errorf("resolve repository root: %w", err)
	}
	policy, err := productownership.Load(bytes.NewReader(current))
	if err != nil {
		return Result{}, fmt.Errorf("load current product ownership policy: %w", err)
	}
	document, err := decodePolicyDocument(current)
	if err != nil {
		return Result{}, err
	}
	index, err := newCatalogIndex()
	if err != nil {
		return Result{}, err
	}
	metadata, exists := index.metadataByName[productName]
	if !exists {
		return Result{}, fmt.Errorf(
			"unknown Provider product %q; choose one of: %s",
			productName,
			strings.Join(productcatalog.Names(), ", "),
		)
	}
	if metadata.MasterData.EnSampleName == "" || metadata.MasterData.Key == "" {
		return Result{}, fmt.Errorf("product %q has incomplete product master data", productName)
	}
	if err := validateProductDirectory(root, productName); err != nil {
		return Result{}, err
	}
	users, err := cleanGitHubUsers(githubUsers)
	if err != nil {
		return Result{}, err
	}

	paths, assets, err := derivePaths(root, productName, policy, metadata, index)
	if err != nil {
		return Result{}, err
	}
	generated := productownership.Product{GitHubUsers: users, Paths: paths}

	var previous productownership.Product
	previousRaw, hadPrevious := document.Products[productName]
	if hadPrevious {
		if err := json.Unmarshal(previousRaw, &previous); err != nil {
			return Result{}, fmt.Errorf("decode current product %q policy: %w", productName, err)
		}
	}
	generatedRaw, err := json.Marshal(generated)
	if err != nil {
		return Result{}, fmt.Errorf("encode generated product %q policy: %w", productName, err)
	}
	document.Products[productName] = generatedRaw
	next, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("encode generated product ownership policy: %w", err)
	}
	next = append(next, '\n')
	generatedPolicy, err := productownership.Load(bytes.NewReader(next))
	if err != nil {
		return Result{}, fmt.Errorf("validate generated product ownership policy: %w", err)
	}
	if err := validateAssetOwnership(generatedPolicy, productName, assets); err != nil {
		return Result{}, err
	}

	added, removed := diffStrings(previous.Paths, generated.Paths)
	return Result{
		ProductName:    productName,
		MasterData:     metadata.MasterData,
		Previous:       previous,
		HadPrevious:    hadPrevious,
		Generated:      generated,
		AddedPaths:     added,
		RemovedPaths:   removed,
		UsersChanged:   !equalStrings(previous.GitHubUsers, generated.GitHubUsers),
		Changed:        !bytes.Equal(current, next),
		Baseline:       append([]byte(nil), current...),
		PolicyContents: next,
	}, nil
}

// ValidateProposal verifies a product-team onboarding policy against a trusted
// baseline. Exactly one product may change, and its entry must equal a fresh
// derivation from the trusted catalog and repository contents.
func ValidateProposal(root string, baseline, proposal []byte, author, sender string) (string, error) {
	baselinePolicy, err := productownership.Load(bytes.NewReader(baseline))
	if err != nil {
		return "", fmt.Errorf("load trusted product ownership policy: %w", err)
	}
	proposalPolicy, err := productownership.Load(bytes.NewReader(proposal))
	if err != nil {
		return "", fmt.Errorf("load proposed product ownership policy: %w", err)
	}
	if baselinePolicy.Version != proposalPolicy.Version {
		return "", fmt.Errorf("proposed policy changes the policy version")
	}
	if !reflect.DeepEqual(baselinePolicy.Core, proposalPolicy.Core) {
		return "", fmt.Errorf("proposed policy changes core GitHub users")
	}

	productNames := make(map[string]struct{}, len(baselinePolicy.Products)+len(proposalPolicy.Products))
	for productName := range baselinePolicy.Products {
		productNames[productName] = struct{}{}
	}
	for productName := range proposalPolicy.Products {
		productNames[productName] = struct{}{}
	}
	var changedProducts []string
	for productName := range productNames {
		baselineProduct, existed := baselinePolicy.Products[productName]
		proposedProduct, exists := proposalPolicy.Products[productName]
		if existed && !exists {
			return "", fmt.Errorf("proposed policy removes product %q", productName)
		}
		if !existed || !reflect.DeepEqual(baselineProduct, proposedProduct) {
			changedProducts = append(changedProducts, productName)
		}
	}
	sort.Strings(changedProducts)
	if len(changedProducts) != 1 {
		return "", fmt.Errorf("proposed policy must change exactly one product; changed products: %s", strings.Join(changedProducts, ", "))
	}
	productName := changedProducts[0]
	proposedProduct := proposalPolicy.Products[productName]
	if !containsGitHubUser(proposedProduct.GitHubUsers, author) {
		return "", fmt.Errorf("pull request author %q must be a proposed owner of product %q", author, productName)
	}
	if !containsGitHubUser(proposedProduct.GitHubUsers, sender) && !containsGitHubUser(baselinePolicy.Core.GitHubUsers, sender) {
		return "", fmt.Errorf("pull request event sender %q is neither a proposed product owner nor a core maintainer", sender)
	}

	expected, err := GenerateFromPolicy(root, baseline, productName, proposedProduct.GitHubUsers)
	if err != nil {
		return "", fmt.Errorf("recompute product %q ownership: %w", productName, err)
	}
	expectedPolicy, err := productownership.Load(bytes.NewReader(expected.PolicyContents))
	if err != nil {
		return "", fmt.Errorf("load recomputed product %q ownership: %w", productName, err)
	}
	if !reflect.DeepEqual(expectedPolicy.Products[productName], proposedProduct) {
		return "", fmt.Errorf(
			"proposed product %q ownership does not match trusted generation; expected users=%v paths=%v",
			productName,
			expectedPolicy.Products[productName].GitHubUsers,
			expectedPolicy.Products[productName].Paths,
		)
	}
	return productName, nil
}

// Write atomically replaces the repository policy after validating its target.
func Write(root string, baseline, contents []byte) error {
	if _, err := productownership.Load(bytes.NewReader(contents)); err != nil {
		return fmt.Errorf("refuse to write invalid product ownership policy: %w", err)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	directory := filepath.Join(root, ".github")
	directoryInfo, err := os.Lstat(directory)
	if err != nil {
		return fmt.Errorf("inspect product ownership policy directory: %w", err)
	}
	if !directoryInfo.IsDir() || directoryInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("product ownership policy directory must be a real directory")
	}
	filename := filepath.Join(directory, "product-owners.json")
	fileInfo, err := os.Lstat(filename)
	if err != nil {
		return fmt.Errorf("inspect product ownership policy: %w", err)
	}
	if !fileInfo.Mode().IsRegular() || fileInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("product ownership policy must be a regular file")
	}
	current, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read product ownership policy before write: %w", err)
	}
	if !bytes.Equal(current, baseline) {
		return fmt.Errorf("product ownership policy changed after generation; rerun the command")
	}

	temporary, err := os.CreateTemp(directory, ".product-owners-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary product ownership policy: %w", err)
	}
	temporaryName := temporary.Name()
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(fileInfo.Mode().Perm()); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set temporary product ownership policy mode: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary product ownership policy: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary product ownership policy: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary product ownership policy: %w", err)
	}
	current, err = os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read product ownership policy before replace: %w", err)
	}
	if !bytes.Equal(current, baseline) {
		return fmt.Errorf("product ownership policy changed during generation; rerun the command")
	}
	if err := os.Rename(temporaryName, filename); err != nil {
		return fmt.Errorf("replace product ownership policy: %w", err)
	}
	removeTemporary = false
	return nil
}

func decodePolicyDocument(contents []byte) (policyDocument, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var document policyDocument
	if err := decoder.Decode(&document); err != nil {
		return policyDocument{}, fmt.Errorf("decode product ownership policy document: %w", err)
	}
	if document.Products == nil {
		return policyDocument{}, fmt.Errorf("product ownership policy document has no products")
	}
	return document, nil
}

func newCatalogIndex() (catalogIndex, error) {
	all, err := productcatalog.AllOwnershipMetadata()
	if err != nil {
		return catalogIndex{}, fmt.Errorf("load product catalog ownership metadata: %w", err)
	}
	index := catalogIndex{
		metadataByName: make(map[string]productcatalog.OwnershipMetadata, len(all)),
		ownerByType:    make(map[string]string),
	}
	for _, metadata := range all {
		index.metadataByName[metadata.Name] = metadata
		for _, terraformType := range append(append([]string(nil), metadata.ResourceTypes...), metadata.DataSourceTypes...) {
			if owner, exists := index.ownerByType[terraformType]; exists && owner != metadata.Name {
				return catalogIndex{}, fmt.Errorf(
					"Terraform type %q is registered by products %q and %q",
					terraformType,
					owner,
					metadata.Name,
				)
			}
			index.ownerByType[terraformType] = metadata.Name
		}
	}
	return index, nil
}

func validateProductDirectory(root, productName string) error {
	filename := filepath.Join(root, "products", productName)
	info, err := os.Lstat(filename)
	if err != nil {
		return fmt.Errorf("inspect product directory %q: %w", filepath.ToSlash(filename), err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("product %q must have a real directory at products/%s", productName, productName)
	}
	return nil
}

func cleanGitHubUsers(users []string) ([]string, error) {
	if len(users) == 0 {
		return nil, fmt.Errorf("at least one -github-user is required")
	}
	cleaned := make([]string, 0, len(users))
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			return nil, fmt.Errorf("GitHub user must not be empty")
		}
		cleaned = append(cleaned, user)
	}
	sort.SliceStable(cleaned, func(i, j int) bool {
		left := strings.ToLower(cleaned[i])
		right := strings.ToLower(cleaned[j])
		if left == right {
			return cleaned[i] < cleaned[j]
		}
		return left < right
	})
	return cleaned, nil
}

func containsGitHubUser(users []string, candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	for _, user := range users {
		if strings.EqualFold(user, candidate) {
			return true
		}
	}
	return false
}

func derivePaths(
	root string,
	productName string,
	policy *productownership.Policy,
	metadata productcatalog.OwnershipMetadata,
	index catalogIndex,
) ([]string, []string, error) {
	examplePaths, exampleAssets, err := deriveExamplePaths(root, productName, policy, index)
	if err != nil {
		return nil, nil, err
	}
	docPaths, docAssets, err := deriveDocPaths(root, productName, policy, metadata, index)
	if err != nil {
		return nil, nil, err
	}
	supplemental := append(examplePaths, docPaths...)
	sort.Strings(supplemental)
	paths := []string{"products/" + productName + "/**"}
	paths = append(paths, uniqueStrings(supplemental)...)
	assets := append(exampleAssets, docAssets...)
	sort.Strings(assets)
	return paths, uniqueStrings(assets), nil
}

func deriveExamplePaths(
	root string,
	productName string,
	policy *productownership.Policy,
	index catalogIndex,
) ([]string, []string, error) {
	examplesRoot := filepath.Join(root, "examples")
	entries, err := os.ReadDir(examplesRoot)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read examples directory: %w", err)
	}
	var ownedPaths []string
	var allFiles []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		directory := filepath.Join(examplesRoot, entry.Name())
		types, files, err := terraformTypesInDirectory(root, directory)
		if err != nil {
			return nil, nil, fmt.Errorf("inspect example %q: %w", entry.Name(), err)
		}
		allFiles = append(allFiles, files...)
		currentOwner, hasCurrentOwner, err := currentOwnerForFiles(policy, files)
		if err != nil {
			return nil, nil, fmt.Errorf("inspect example %q ownership: %w", entry.Name(), err)
		}
		referencedProducts := make(map[string]struct{})
		for _, terraformType := range types {
			owner, exists := index.ownerByType[terraformType]
			if !exists {
				return nil, nil, fmt.Errorf("example %q uses unregistered Terraform type %q", entry.Name(), terraformType)
			}
			referencedProducts[owner] = struct{}{}
		}
		_, referencesTarget := referencedProducts[productName]
		if hasCurrentOwner && currentOwner == productName && !referencesTarget {
			return nil, nil, fmt.Errorf(
				"example %q is owned by product %q but does not use one of its Terraform types",
				entry.Name(),
				productName,
			)
		}
		if referencesTarget && (!hasCurrentOwner || currentOwner == productName) {
			ownedPaths = append(ownedPaths, "examples/"+entry.Name()+"/**")
		}
	}
	return ownedPaths, allFiles, nil
}

func terraformTypesInDirectory(root, directory string) ([]string, []string, error) {
	parser := hclparse.NewParser()
	types := make(map[string]struct{})
	var files []string
	err := filepath.WalkDir(directory, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if filename != directory && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("example contains symbolic link %q", entry.Name())
		}
		relative, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		if filepath.Ext(filename) != ".tf" {
			return nil
		}
		file, diagnostics := parser.ParseHCLFile(filename)
		if diagnostics.HasErrors() {
			return fmt.Errorf("parse %s: %s", filepath.ToSlash(relative), diagnostics.Error())
		}
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			return fmt.Errorf("parse %s: unsupported Terraform body", filepath.ToSlash(relative))
		}
		for _, block := range body.Blocks {
			if (block.Type == "resource" || block.Type == "data") && len(block.Labels) >= 1 && strings.HasPrefix(block.Labels[0], "ucloud_") {
				types[block.Labels[0]] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return sortedKeys(types), files, nil
}

func deriveDocPaths(
	root string,
	productName string,
	policy *productownership.Policy,
	metadata productcatalog.OwnershipMetadata,
	index catalogIndex,
) ([]string, []string, error) {
	assets, err := scanDocAssets(root, policy, index)
	if err != nil {
		return nil, nil, err
	}
	var targetAssets []docAsset
	var assetPaths []string
	for _, asset := range assets {
		assetPaths = append(assetPaths, asset.path)
		if asset.resolvedOwner == productName {
			targetAssets = append(targetAssets, asset)
		}
	}

	var patterns []string
	if existing, exists := policy.Products[productName]; exists {
		for _, pattern := range existing.Paths {
			if !strings.HasPrefix(pattern, "website/docs/") {
				continue
			}
			matched := false
			safe := true
			for _, asset := range assets {
				if !productownership.MatchesPath(pattern, asset.path) {
					continue
				}
				matched = true
				if asset.resolvedOwner != productName {
					safe = false
					break
				}
			}
			if matched && safe {
				patterns = append(patterns, pattern)
			}
		}
	}

	candidates := docPatternCandidates(metadata, targetAssets)
	for _, asset := range targetAssets {
		if matchesAny(patterns, asset.path) {
			continue
		}
		selected := ""
		for _, candidate := range candidates {
			if !productownership.MatchesPath(candidate, asset.path) || !docPatternSafe(candidate, productName, assets) {
				continue
			}
			selected = candidate
			break
		}
		if selected == "" {
			selected = asset.path
		}
		patterns = append(patterns, selected)
	}
	sort.Strings(patterns)
	return uniqueStrings(patterns), assetPaths, nil
}

func scanDocAssets(root string, policy *productownership.Policy, index catalogIndex) ([]docAsset, error) {
	docsRoot := filepath.Join(root, "website", "docs")
	var assets []docAsset
	err := filepath.WalkDir(docsRoot, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("website docs contain symbolic link %q", entry.Name())
		}
		relative, err := filepath.Rel(root, filename)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		parts := strings.Split(relative, "/")
		if len(parts) < 4 || parts[0] != "website" || parts[1] != "docs" {
			return nil
		}
		section := parts[2]
		stem := docStem(entry.Name())
		inferredOwner := ""
		switch section {
		case "r", "d":
			inferredOwner = index.ownerByType["ucloud_"+stem]
		case "appendix":
			inferredOwner, err = inferAppendixOwner(stem, index.metadataByName)
			if err != nil {
				return err
			}
		}
		currentOwner, hasCurrentOwner, err := policy.ProductOwner(relative)
		if err != nil {
			return err
		}
		if hasCurrentOwner && inferredOwner != "" && currentOwner != inferredOwner {
			return fmt.Errorf(
				"website document %q is owned by product %q but its Terraform type belongs to %q",
				relative,
				currentOwner,
				inferredOwner,
			)
		}
		resolvedOwner := inferredOwner
		if hasCurrentOwner {
			resolvedOwner = currentOwner
		}
		assets = append(assets, docAsset{
			path:          relative,
			section:       section,
			stem:          stem,
			resolvedOwner: resolvedOwner,
		})
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan website docs: %w", err)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].path < assets[j].path })
	return assets, nil
}

func inferAppendixOwner(
	stem string,
	metadataByName map[string]productcatalog.OwnershipMetadata,
) (string, error) {
	owner := ""
	matchedLength := 0
	for productName, metadata := range metadataByName {
		for _, namespace := range metadata.TerraformNamespaces {
			if !stemMatchesNamespace(stem, namespace) {
				continue
			}
			if len(namespace) > matchedLength {
				owner = productName
				matchedLength = len(namespace)
				continue
			}
			if len(namespace) == matchedLength && owner != productName {
				return "", fmt.Errorf("appendix document %q matches products %q and %q", stem, owner, productName)
			}
		}
	}
	return owner, nil
}

func docPatternCandidates(metadata productcatalog.OwnershipMetadata, assets []docAsset) []string {
	var candidates []string
	for _, section := range []string{"d", "r"} {
		for _, namespace := range metadata.TerraformNamespaces {
			var stems []string
			for _, asset := range assets {
				if asset.section == section && stemMatchesNamespace(asset.stem, namespace) {
					stems = append(stems, asset.stem)
				}
			}
			if len(stems) == 0 {
				continue
			}
			prefix := namespace + "_"
			for _, stem := range stems {
				if stem == namespace || stem == namespace+"s" {
					prefix = namespace
					break
				}
			}
			candidates = append(candidates, "website/docs/"+section+"/"+prefix+"*")
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if len(candidates[i]) == len(candidates[j]) {
			return candidates[i] < candidates[j]
		}
		return len(candidates[i]) < len(candidates[j])
	})
	return uniqueStrings(candidates)
}

func docPatternSafe(pattern, productName string, assets []docAsset) bool {
	matched := false
	for _, asset := range assets {
		if !productownership.MatchesPath(pattern, asset.path) {
			continue
		}
		matched = true
		if asset.resolvedOwner != productName {
			return false
		}
	}
	return matched
}

func currentOwnerForFiles(policy *productownership.Policy, files []string) (string, bool, error) {
	owner := ""
	found := false
	for _, filename := range files {
		candidate, owned, err := policy.ProductOwner(filename)
		if err != nil {
			return "", false, err
		}
		if !owned {
			continue
		}
		if found && owner != candidate {
			return "", false, fmt.Errorf("files are assigned to products %q and %q", owner, candidate)
		}
		owner = candidate
		found = true
	}
	return owner, found, nil
}

func validateAssetOwnership(policy *productownership.Policy, productName string, assetPaths []string) error {
	for _, filename := range assetPaths {
		if _, _, err := policy.ProductOwner(filename); err != nil {
			return fmt.Errorf("validate generated ownership for %q: %w", filename, err)
		}
	}
	product, exists := policy.Products[productName]
	if !exists {
		return fmt.Errorf("generated policy is missing product %q", productName)
	}
	for _, pattern := range product.Paths {
		if strings.HasPrefix(pattern, "products/") {
			continue
		}
		if !matchesAnyPath(pattern, assetPaths) {
			return fmt.Errorf("generated product %q path %q does not match a repository asset", productName, pattern)
		}
	}
	return nil
}

func docStem(filename string) string {
	for _, suffix := range []string{".html.markdown", ".markdown", ".md"} {
		if strings.HasSuffix(filename, suffix) {
			return strings.TrimSuffix(filename, suffix)
		}
	}
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

func stemMatchesNamespace(stem, namespace string) bool {
	return stem == namespace ||
		strings.HasPrefix(stem, namespace+"_") ||
		stem == namespace+"s" ||
		strings.HasPrefix(stem, namespace+"s_")
}

func matchesAny(patterns []string, filename string) bool {
	for _, pattern := range patterns {
		if productownership.MatchesPath(pattern, filename) {
			return true
		}
	}
	return false
}

func matchesAnyPath(pattern string, filenames []string) bool {
	for _, filename := range filenames {
		if productownership.MatchesPath(pattern, filename) {
			return true
		}
	}
	return false
}

func diffStrings(previous, generated []string) (added, removed []string) {
	previousSet := make(map[string]struct{}, len(previous))
	generatedSet := make(map[string]struct{}, len(generated))
	for _, value := range previous {
		previousSet[value] = struct{}{}
	}
	for _, value := range generated {
		generatedSet[value] = struct{}{}
		if _, exists := previousSet[value]; !exists {
			added = append(added, value)
		}
	}
	for _, value := range previous {
		if _, exists := generatedSet[value]; !exists {
			removed = append(removed, value)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
