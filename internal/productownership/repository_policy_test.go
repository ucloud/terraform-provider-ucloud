package productownership_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

func TestRepositoryProductOwnershipPolicyLoads(t *testing.T) {
	file, err := os.Open("../../.github/product-owners.json")
	if err != nil {
		t.Fatalf("open repository product ownership policy: %v", err)
	}
	defer file.Close()

	policy, err := productownership.Load(file)
	if err != nil {
		t.Fatalf("load repository product ownership policy: %v", err)
	}

	productsRoot := filepath.Join("..", "..", "products")
	entries, err := os.ReadDir(productsRoot)
	if err != nil {
		t.Fatalf("read product packages: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if _, exists := policy.Products[entry.Name()]; !exists {
				t.Errorf("product package %q has no ownership policy", entry.Name())
			}
		}
	}
	for name := range policy.Products {
		info, err := os.Stat(filepath.Join(productsRoot, name))
		if err != nil {
			t.Errorf("ownership policy product %q has no package directory: %v", name, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("ownership policy product %q is not a package directory", name)
		}
	}
}

func TestRepositoryExamplesAndDocsHaveExplicitOwnership(t *testing.T) {
	root := repositoryRoot(t)
	policy := loadRepositoryPolicy(t, root)
	policy.Core.GitHubUsers = []string{"coverage-core"}
	for name, product := range policy.Products {
		product.GitHubUsers = []string{"coverage-owner"}
		policy.Products[name] = product
	}

	for _, scope := range []string{"examples", filepath.Join("website", "docs")} {
		err := filepath.WalkDir(filepath.Join(root, scope), func(filename string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			relative, err := filepath.Rel(root, filename)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			if relative == "examples/README.md" ||
				relative == "website/docs/index.html.markdown" {
				return nil
			}

			decision, err := policy.Authorize("coverage-owner", []productownership.Change{{Path: relative}})
			if err != nil {
				t.Errorf("%s must have exactly one product owner: %v", relative, err)
				return nil
			}
			if decision.Owner == "" || decision.Owner == "core" {
				t.Errorf("%s resolved to invalid product owner %q", relative, decision.Owner)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s ownership scope: %v", scope, err)
		}
	}
}

func TestProductOwnershipWorkflowNeverExecutesPullRequestCode(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/product-ownership.yml")
	if err != nil {
		t.Fatalf("read product ownership workflow: %v", err)
	}
	content := string(workflow)
	if !strings.Contains(content, "pull_request_target:") {
		t.Fatal("product ownership workflow must run from pull_request_target")
	}
	for _, forbidden := range []string{
		"github.event.pull_request.head",
		"allow-unsafe-pr-checkout",
		"gh pr checkout",
		"git fetch",
		"uses: actions/checkout@v",
		"uses: actions/setup-go@v",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("product ownership workflow contains unsafe pull request checkout token %q", forbidden)
		}
	}
	if !strings.Contains(content, "statuses: write") {
		t.Fatal("product ownership workflow must publish the required status on the pull request head")
	}
	if !strings.Contains(content, "- edited") {
		t.Fatal("product ownership workflow must recheck pull requests after metadata edits")
	}
}

func TestProductAcceptanceWorkflowUsesProductEnvironmentOnMaster(t *testing.T) {
	workflow, err := os.ReadFile("../../.github/workflows/product-acceptance.yml")
	if err != nil {
		t.Fatalf("read product acceptance workflow: %v", err)
	}
	content := string(workflow)
	if strings.Count(content, "github.ref == 'refs/heads/master'") != 2 {
		t.Fatal("product acceptance entry and upgrade jobs must be restricted to the default branch")
	}
	if !strings.Contains(content, "acceptance_environment:") || !strings.Contains(content, "type: environment") {
		t.Fatal("product acceptance workflow must use an environment input")
	}
	if strings.Contains(content, "type: choice") || strings.Contains(content, "        options:") {
		t.Fatal("product acceptance workflow must not maintain a static product option list")
	}
	if strings.Count(content, "name: ${{ needs.resolve.outputs.environment }}") != 2 {
		t.Fatal("product acceptance jobs must use the validated selected environment")
	}
	for _, forbidden := range []string{
		"uses: actions/checkout@v",
		"uses: actions/setup-go@v",
		"uses: hashicorp/setup-terraform@v",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("product acceptance workflow contains mutable action reference %q", forbidden)
		}
	}
	if !strings.Contains(content, "PRODUCT: ${{ needs.resolve.outputs.product }}") ||
		!strings.Contains(content, `go test "./products/${PRODUCT}"`) {
		t.Fatal("product acceptance workflow must test the selected product package")
	}
	if !strings.Contains(content, `^product-([a-z][a-z0-9_-]*)-acceptance$`) ||
		!strings.Contains(content, `[[ ! -d "products/${product}" ]]`) {
		t.Fatal("product acceptance workflow must validate the environment name and product directory before loading secrets")
	}
	if !strings.Contains(content, `-list '^TestAcc'`) {
		t.Fatal("product acceptance workflow must reject packages without acceptance tests")
	}
	if !strings.Contains(content, "needs.resolve.outputs.product == 'us3'") {
		t.Fatal("the released-to-current upgrade fixture must remain restricted to US3")
	}
	if strings.Count(content, "name: Mask cloud credentials") != 2 ||
		strings.Count(content, `TF_LOG: "OFF"`) != 2 {
		t.Fatal("acceptance and upgrade jobs must mask credentials and disable Terraform debug logs")
	}
	if !strings.Contains(content, "public_key|private_key|security_token") {
		t.Fatal("acceptance log filtering must cover common snake_case credential fields")
	}
}

func TestMakefileDiscoversProductDirectories(t *testing.T) {
	makefile, err := os.ReadFile("../../GNUmakefile")
	if err != nil {
		t.Fatalf("read GNUmakefile: %v", err)
	}
	content := string(makefile)
	if !strings.Contains(content, "PRODUCT_DIRS:=$(wildcard products/*/)") ||
		!strings.Contains(content, "PRODUCTS:=$(sort $(notdir $(patsubst %/,%,$(PRODUCT_DIRS))))") {
		t.Fatal("GNUmakefile must derive its product list from products/*/")
	}
}
