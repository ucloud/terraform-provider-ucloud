package productownership

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
)

const policyVersion = 1

var productNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// Policy is the trusted mapping between GitHub users and product-owned paths.
type Policy struct {
	Version  int                `json:"version"`
	Core     Owners             `json:"core"`
	Products map[string]Product `json:"products"`
}

// Owners identifies GitHub accounts with repository-wide ownership.
type Owners struct {
	GitHubUsers []string `json:"github_users"`
}

// Product identifies the GitHub accounts and paths owned by one product.
type Product struct {
	GitHubUsers []string `json:"github_users"`
	Paths       []string `json:"paths"`
}

// Change represents one path reported by the GitHub pull request files API.
// PreviousPath is populated for renames so both ownership boundaries are checked.
type Change struct {
	Path         string
	PreviousPath string
}

// Decision describes the ownership boundary that authorized a pull request.
type Decision struct {
	Owner string
}

// Load parses and validates a product ownership policy.
func Load(reader io.Reader) (*Policy, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return nil, fmt.Errorf("decode product ownership policy: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}
	if policy.Version != policyVersion {
		return nil, fmt.Errorf("product ownership policy version is %d, want %d", policy.Version, policyVersion)
	}
	if len(policy.Products) == 0 {
		return nil, fmt.Errorf("product ownership policy has no products")
	}

	var err error
	policy.Core.GitHubUsers, err = normalizeUsers(policy.Core.GitHubUsers)
	if err != nil {
		return nil, fmt.Errorf("core GitHub users: %w", err)
	}
	if len(policy.Core.GitHubUsers) == 0 {
		return nil, fmt.Errorf("product ownership policy has no core GitHub users")
	}
	pathOwners := make(map[string]string)
	for name, product := range policy.Products {
		if !productNamePattern.MatchString(name) {
			return nil, fmt.Errorf("product name %q must match %s", name, productNamePattern)
		}
		product.GitHubUsers, err = normalizeUsers(product.GitHubUsers)
		if err != nil {
			return nil, fmt.Errorf("product %q GitHub users: %w", name, err)
		}
		if len(product.GitHubUsers) == 0 {
			return nil, fmt.Errorf("product %q has no GitHub users", name)
		}
		if len(product.Paths) == 0 {
			return nil, fmt.Errorf("product %q has no owned paths", name)
		}
		primaryPath := "products/" + name + "/**"
		hasPrimaryPath := false
		for _, pattern := range product.Paths {
			if err := validatePattern(pattern); err != nil {
				return nil, fmt.Errorf("product %q path %q: %w", name, pattern, err)
			}
			if err := validateProductPattern(name, pattern); err != nil {
				return nil, fmt.Errorf("product %q path %q: %w", name, pattern, err)
			}
			if existing, exists := pathOwners[pattern]; exists {
				if existing != name {
					return nil, fmt.Errorf("path %q is assigned to multiple products: %q and %q", pattern, existing, name)
				}
				return nil, fmt.Errorf("product %q contains duplicate path %q", name, pattern)
			}
			pathOwners[pattern] = name
			if pattern == primaryPath {
				hasPrimaryPath = true
			}
		}
		if !hasPrimaryPath {
			return nil, fmt.Errorf("product %q must own primary path %q", name, primaryPath)
		}
		policy.Products[name] = product
	}
	return &policy, nil
}

// Authorize verifies that one GitHub user owns every changed path.
func (policy *Policy) Authorize(author string, changes []Change) (Decision, error) {
	if policy == nil {
		return Decision{}, fmt.Errorf("product ownership policy is nil")
	}
	author = normalizeUser(author)
	if author == "" {
		return Decision{}, fmt.Errorf("pull request author is empty")
	}
	if len(changes) == 0 {
		return Decision{}, fmt.Errorf("pull request has no changed files")
	}
	for _, change := range changes {
		for _, changedPath := range []string{change.Path, change.PreviousPath} {
			if changedPath == "" {
				continue
			}
			if err := validateRepositoryPath(changedPath); err != nil {
				return Decision{}, fmt.Errorf("changed path %q: %w", changedPath, err)
			}
		}
	}
	if contains(policy.Core.GitHubUsers, author) {
		return Decision{Owner: "core"}, nil
	}

	owner := ""
	for _, change := range changes {
		paths := []string{change.Path}
		if change.PreviousPath != "" {
			paths = append(paths, change.PreviousPath)
		}
		for _, changedPath := range paths {
			productName, err := policy.ownerOf(changedPath)
			if err != nil {
				return Decision{}, err
			}
			if owner == "" {
				owner = productName
			} else if owner != productName {
				return Decision{}, fmt.Errorf("pull request changes multiple products: %q and %q", owner, productName)
			}
		}
	}

	product := policy.Products[owner]
	if !contains(product.GitHubUsers, author) {
		return Decision{}, fmt.Errorf("GitHub user %q is not allowed to change product %q", author, owner)
	}
	return Decision{Owner: owner}, nil
}

func (policy *Policy) ownerOf(changedPath string) (string, error) {
	if err := validateRepositoryPath(changedPath); err != nil {
		return "", fmt.Errorf("changed path %q: %w", changedPath, err)
	}

	owners := make([]string, 0, 1)
	for productName, product := range policy.Products {
		for _, pattern := range product.Paths {
			if matchPath(pattern, changedPath) {
				owners = append(owners, productName)
				break
			}
		}
	}
	sort.Strings(owners)
	switch len(owners) {
	case 0:
		return "", fmt.Errorf("path %q is not owned by a product and requires a core maintainer", changedPath)
	case 1:
		return owners[0], nil
	default:
		return "", fmt.Errorf("path %q matches multiple products: %s", changedPath, strings.Join(owners, ", "))
	}
}

func matchPath(pattern, name string) bool {
	patternParts := strings.Split(pattern, "/")
	nameParts := strings.Split(name, "/")
	return matchParts(patternParts, nameParts)
}

func matchParts(pattern, name []string) bool {
	if len(pattern) == 0 {
		return len(name) == 0
	}
	if pattern[0] == "**" {
		if matchParts(pattern[1:], name) {
			return true
		}
		return len(name) > 0 && matchParts(pattern, name[1:])
	}
	if len(name) == 0 {
		return false
	}
	matched, err := path.Match(pattern[0], name[0])
	return err == nil && matched && matchParts(pattern[1:], name[1:])
}

func validatePattern(pattern string) error {
	if err := validateRepositoryPath(pattern); err != nil {
		return err
	}
	for _, part := range strings.Split(pattern, "/") {
		if part == "**" {
			continue
		}
		if _, err := path.Match(part, "candidate"); err != nil {
			return fmt.Errorf("invalid glob: %w", err)
		}
	}
	return nil
}

func validateProductPattern(productName, pattern string) error {
	parts := strings.Split(pattern, "/")
	switch parts[0] {
	case "products":
		if len(parts) < 3 || parts[1] != productName {
			return fmt.Errorf("products path must stay under products/%s", productName)
		}
	case "examples":
		if len(parts) < 2 || strings.ContainsAny(parts[1], "*?[") {
			return fmt.Errorf("examples path must identify one specific product-owned entry")
		}
	case "website":
		if len(parts) < 4 || parts[1] != "docs" ||
			strings.ContainsAny(parts[2], "*?[") || strings.ContainsAny(parts[3][:1], "*?[") {
			return fmt.Errorf("website/docs path must identify one specific product-owned entry")
		}
	default:
		return fmt.Errorf("path is core-owned; product paths are limited to products, examples, and website/docs")
	}
	return nil
}

func validateRepositoryPath(name string) error {
	if name == "" {
		return fmt.Errorf("path is empty")
	}
	if strings.Contains(name, "\\") {
		return fmt.Errorf("backslashes are not allowed")
	}
	if strings.HasPrefix(name, "/") || path.Clean(name) != name {
		return fmt.Errorf("path must be a clean repository-relative path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." || part == "." || part == "" {
			return fmt.Errorf("path must not contain traversal or empty segments")
		}
	}
	return nil
}

func normalizeUsers(users []string) ([]string, error) {
	normalized := make([]string, 0, len(users))
	seen := make(map[string]struct{}, len(users))
	for _, user := range users {
		user = normalizeUser(user)
		if user == "" {
			continue
		}
		if !validGitHubUser(user) {
			return nil, fmt.Errorf("invalid GitHub user %q", user)
		}
		if _, exists := seen[user]; exists {
			return nil, fmt.Errorf("duplicate GitHub user %q", user)
		}
		seen[user] = struct{}{}
		normalized = append(normalized, user)
	}
	return normalized, nil
}

func normalizeUser(user string) string {
	return strings.ToLower(strings.TrimSpace(user))
}

func validGitHubUser(user string) bool {
	if len(user) < 1 || len(user) > 39 || user[0] == '-' || user[len(user)-1] == '-' {
		return false
	}
	previousHyphen := false
	for _, char := range user {
		if char == '-' {
			if previousHyphen {
				return false
			}
			previousHyphen = true
			continue
		}
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return false
		}
		previousHyphen = false
	}
	return true
}

func contains(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode product ownership policy: multiple JSON values")
		}
		return fmt.Errorf("decode product ownership policy: %w", err)
	}
	return nil
}
