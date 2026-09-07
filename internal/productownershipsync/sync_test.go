package productownershipsync_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownershipsync"
)

func TestGenerateDerivesPathsAndPreservesHistoricalAssignments(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{
		policy: `{
  "version": 1,
  "core": {"github_users": ["CoreOwner"]},
  "products": {
    "uhost": {
      "github_users": ["UHostOwner"],
      "paths": ["products/uhost/**", "examples/web/**", "website/docs/r/instance*"]
    },
    "ulb": {
      "github_users": ["OldOwner"],
      "paths": ["products/ulb/**", "website/docs/appendix/load_balancer_notes.markdown", "website/docs/r/lb*"]
    }
  }
}
`,
		files: map[string]string{
			"examples/shared/main.tf": `resource "ucloud_lb" "main" {}
resource "ucloud_instance" "backend" {}
`,
			"examples/web/main.tf":                               `resource "ucloud_instance" "web" {}`,
			"website/docs/appendix/load_balancer_notes.markdown": "historical ULB notes\n",
			"website/docs/d/lbs.html.markdown":                   "---\npage_title: ULBs\n---\n",
			"website/docs/r/instance.html.markdown":              "---\npage_title: Instance\n---\n",
			"website/docs/r/lb.html.markdown":                    "---\npage_title: Load Balancer\n---\n",
			"website/docs/r/lb_listener.html.markdown":           "---\npage_title: Listener\n---\n",
		},
	})
	policyPath := filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath))
	before, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read fixture policy: %v", err)
	}

	result, err := productownershipsync.Generate(root, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	wantPaths := []string{
		"products/ulb/**",
		"examples/shared/**",
		"website/docs/appendix/load_balancer_notes.markdown",
		"website/docs/d/lb*",
		"website/docs/r/lb*",
	}
	if !reflect.DeepEqual(result.Generated.Paths, wantPaths) {
		t.Fatalf("generated paths = %v, want %v", result.Generated.Paths, wantPaths)
	}
	if got, want := result.Generated.GitHubUsers, []string{"NewOwner"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("generated GitHub users = %v, want %v", got, want)
	}
	if !result.Changed || !result.UsersChanged {
		t.Fatalf("result changes = changed:%t users:%t, want both true", result.Changed, result.UsersChanged)
	}
	if got, want := result.AddedPaths, []string{"examples/shared/**", "website/docs/d/lb*"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("added paths = %v, want %v", got, want)
	}
	if len(result.RemovedPaths) != 0 {
		t.Fatalf("removed paths = %v, want none", result.RemovedPaths)
	}
	afterDryRun, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read policy after Generate: %v", err)
	}
	if !bytes.Equal(before, afterDryRun) {
		t.Fatal("Generate changed the policy during dry-run")
	}

	policy, err := productownership.Load(bytes.NewReader(result.PolicyContents))
	if err != nil {
		t.Fatalf("load generated policy: %v", err)
	}
	if got := policy.Products["uhost"].GitHubUsers; !reflect.DeepEqual(got, []string{"uhostowner"}) {
		t.Fatalf("non-target product changed: %v", got)
	}
}

func TestWriteIsAtomicModePreservingAndIdempotent(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{
		policy: minimalPolicy("OldOwner"),
	})
	policyPath := filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath))
	if err := os.Chmod(policyPath, 0640); err != nil {
		t.Fatalf("chmod fixture policy: %v", err)
	}
	result, err := productownershipsync.Generate(root, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if err := productownershipsync.Write(root, result.Baseline, result.PolicyContents); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	info, err := os.Stat(policyPath)
	if err != nil {
		t.Fatalf("stat written policy: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0640); got != want {
		t.Fatalf("written policy mode = %o, want %o", got, want)
	}

	second, err := productownershipsync.Generate(root, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("second Generate() error = %v", err)
	}
	if second.Changed || second.UsersChanged || len(second.AddedPaths) != 0 || len(second.RemovedPaths) != 0 {
		t.Fatalf("second Generate() is not idempotent: %#v", second)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".github"))
	if err != nil {
		t.Fatalf("read policy directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".product-owners-") {
			t.Fatalf("temporary policy file was not removed: %s", entry.Name())
		}
	}
}

func TestWriteRejectsSymlinkPolicy(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: minimalPolicy("Owner")})
	policyPath := filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath))
	targetPath := filepath.Join(root, "outside-policy.json")
	if err := os.WriteFile(targetPath, []byte(minimalPolicy("Owner")), 0644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Remove(policyPath); err != nil {
		t.Fatalf("remove fixture policy: %v", err)
	}
	if err := os.Symlink(targetPath, policyPath); err != nil {
		t.Skipf("create policy symlink: %v", err)
	}
	before, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read symlink target: %v", err)
	}
	if err := productownershipsync.Write(root, before, before); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("Write(symlink) error = %v, want regular-file error", err)
	}
	after, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read symlink target after Write: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("Write modified a symlink target")
	}
}

func TestWriteRejectsPolicyChangedAfterGeneration(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: minimalPolicy("Owner")})
	result, err := productownershipsync.Generate(root, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	policyPath := filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath))
	concurrent := []byte(minimalPolicy("ConcurrentOwner"))
	if err := os.WriteFile(policyPath, concurrent, 0644); err != nil {
		t.Fatalf("write concurrent policy change: %v", err)
	}
	if err := productownershipsync.Write(root, result.Baseline, result.PolicyContents); err == nil || !strings.Contains(err.Error(), "changed after generation") {
		t.Fatalf("Write(concurrent change) error = %v, want stale-baseline error", err)
	}
	after, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read policy after rejected Write: %v", err)
	}
	if !bytes.Equal(after, concurrent) {
		t.Fatal("Write overwrote a concurrent policy change")
	}
}

func TestGenerateRejectsUnknownProductAndTerraformType(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: minimalPolicy("Owner")})
	if _, err := productownershipsync.Generate(root, "ULB", []string{"Owner"}); err == nil || !strings.Contains(err.Error(), "unknown Provider product") {
		t.Fatalf("Generate(ULB) error = %v, want exact product-name error", err)
	}

	writeFixtureFile(t, root, "examples/future/main.tf", `resource "ucloud_future_widget" "main" {}`)
	if _, err := productownershipsync.Generate(root, "ulb", []string{"Owner"}); err == nil || !strings.Contains(err.Error(), "unregistered Terraform type") {
		t.Fatalf("Generate(unknown type) error = %v, want unregistered type error", err)
	}
}

func TestGenerateRejectsInvalidGitHubUser(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: minimalPolicy("Owner")})
	if _, err := productownershipsync.Generate(root, "ulb", []string{"owner@example.com"}); err == nil || !strings.Contains(err.Error(), "GitHub user") {
		t.Fatalf("Generate(invalid user) error = %v, want GitHub user validation error", err)
	}
}

func TestGenerateAddsCatalogProductMissingFromPolicy(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: `{
  "version": 1,
  "core": {"github_users": ["CoreOwner"]},
  "products": {
    "uhost": {
      "github_users": ["UHostOwner"],
      "paths": ["products/uhost/**"]
    }
  }
}
`})
	result, err := productownershipsync.Generate(root, "ulb", []string{"ULBOwner"})
	if err != nil {
		t.Fatalf("Generate(new product policy) error = %v", err)
	}
	if result.HadPrevious {
		t.Fatal("Generate reported an existing ULB policy")
	}
	if got, want := result.Generated.Paths, []string{"products/ulb/**"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("new ULB paths = %v, want %v", got, want)
	}
	policy, err := productownership.Load(bytes.NewReader(result.PolicyContents))
	if err != nil {
		t.Fatalf("load generated policy: %v", err)
	}
	if _, exists := policy.Products["uhost"]; !exists {
		t.Fatal("Generate removed the existing UHost policy")
	}
	if got := policy.Products["ulb"].GitHubUsers; !reflect.DeepEqual(got, []string{"ulbowner"}) {
		t.Fatalf("new ULB owners = %v, want normalized ulbowner", got)
	}
}

func TestValidateProposalAllowsOneGeneratedProductChange(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: minimalPolicy("OldOwner")})
	baseline, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath)))
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	generated, err := productownershipsync.GenerateFromPolicy(root, baseline, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("GenerateFromPolicy() error = %v", err)
	}
	owner, err := productownershipsync.ValidateProposal(root, baseline, generated.PolicyContents, "NewOwner", "NewOwner")
	if err != nil {
		t.Fatalf("ValidateProposal() error = %v", err)
	}
	if owner != "ulb" {
		t.Fatalf("ValidateProposal() owner = %q, want ulb", owner)
	}
	if _, err := productownershipsync.ValidateProposal(root, baseline, generated.PolicyContents, "NewOwner", "CoreOwner"); err != nil {
		t.Fatalf("ValidateProposal(core sender) error = %v", err)
	}
}

func TestValidateProposalRejectsTamperingAndUnauthorizedActors(t *testing.T) {
	root := newRepositoryFixture(t, repositoryFixture{policy: minimalPolicy("OldOwner")})
	baseline, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath)))
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	generated, err := productownershipsync.GenerateFromPolicy(root, baseline, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("GenerateFromPolicy() error = %v", err)
	}
	var proposal map[string]interface{}
	if err := json.Unmarshal(generated.PolicyContents, &proposal); err != nil {
		t.Fatalf("decode generated proposal: %v", err)
	}
	products := proposal["products"].(map[string]interface{})
	ulb := products["ulb"].(map[string]interface{})
	ulb["paths"] = append(ulb["paths"].([]interface{}), "website/docs/r/lb_listener.html.markdown")
	tampered, err := json.Marshal(proposal)
	if err != nil {
		t.Fatalf("encode tampered proposal: %v", err)
	}
	coreTampered := bytes.Replace(generated.PolicyContents, []byte("CoreOwner"), []byte("OtherCore"), 1)

	tests := map[string]struct {
		proposal []byte
		author   string
		sender   string
		want     string
	}{
		"extra path": {
			proposal: tampered,
			author:   "NewOwner",
			sender:   "NewOwner",
			want:     "does not match trusted generation",
		},
		"author not proposed": {
			proposal: generated.PolicyContents,
			author:   "OtherOwner",
			sender:   "NewOwner",
			want:     "must be a proposed owner",
		},
		"sender unauthorized": {
			proposal: generated.PolicyContents,
			author:   "NewOwner",
			sender:   "OtherOwner",
			want:     "neither a proposed product owner nor a core maintainer",
		},
		"core changed": {
			proposal: coreTampered,
			author:   "NewOwner",
			sender:   "NewOwner",
			want:     "changes core GitHub users",
		},
		"no product changed": {
			proposal: baseline,
			author:   "OldOwner",
			sender:   "OldOwner",
			want:     "must change exactly one product",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := productownershipsync.ValidateProposal(root, baseline, test.proposal, test.author, test.sender)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateProposal() error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestRepositoryPolicyIsReproducibleForEveryProduct(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repository test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(productownershipsync.PolicyRelativePath)))
	if err != nil {
		t.Fatalf("read repository policy: %v", err)
	}
	var document struct {
		Products map[string]productownership.Product `json:"products"`
	}
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("decode repository policy: %v", err)
	}
	for productName, productPolicy := range document.Products {
		productName := productName
		productPolicy := productPolicy
		t.Run(productName, func(t *testing.T) {
			result, err := productownershipsync.GenerateFromPolicy(root, contents, productName, productPolicy.GitHubUsers)
			if err != nil {
				t.Fatalf("GenerateFromPolicy() error = %v", err)
			}
			if result.Changed {
				t.Fatalf(
					"repository policy is not reproducible; added=%v removed=%v users_changed=%t",
					result.AddedPaths,
					result.RemovedPaths,
					result.UsersChanged,
				)
			}
		})
	}
}

type repositoryFixture struct {
	policy string
	files  map[string]string
}

func newRepositoryFixture(t *testing.T, fixture repositoryFixture) string {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, productownershipsync.PolicyRelativePath, fixture.policy)
	for _, productName := range []string{"uhost", "ulb"} {
		if err := os.MkdirAll(filepath.Join(root, "products", productName), 0755); err != nil {
			t.Fatalf("create product directory: %v", err)
		}
	}
	for filename, contents := range fixture.files {
		writeFixtureFile(t, root, filename, contents)
	}
	return root
}

func writeFixtureFile(t *testing.T, root, relative, contents string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		t.Fatalf("create fixture directory for %s: %v", relative, err)
	}
	if err := os.WriteFile(filename, []byte(contents), 0644); err != nil {
		t.Fatalf("write fixture %s: %v", relative, err)
	}
}

func minimalPolicy(owner string) string {
	document := map[string]interface{}{
		"version": 1,
		"core": map[string]interface{}{
			"github_users": []string{"CoreOwner"},
		},
		"products": map[string]interface{}{
			"ulb": map[string]interface{}{
				"github_users": []string{owner},
				"paths":        []string{"products/ulb/**"},
			},
		},
	}
	contents, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(contents) + "\n"
}
