package productownership_test

import (
	"os"
	"path/filepath"
	"sort"
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
		t.Fatal("product acceptance jobs must be restricted to the default branch")
	}
	if strings.Count(content, "name: product-${{ inputs.product }}-acceptance") != 2 {
		t.Fatal("product acceptance jobs must use a product-specific environment")
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
	if !strings.Contains(content, "PRODUCT: ${{ inputs.product }}") ||
		!strings.Contains(content, `go test "./products/${PRODUCT}"`) {
		t.Fatal("product acceptance workflow must test the selected product package")
	}
	if !strings.Contains(content, `-list '^TestAcc'`) {
		t.Fatal("product acceptance workflow must reject packages without acceptance tests")
	}
	if !strings.Contains(content, "inputs.product == 'us3'") {
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

func TestProductAcceptanceWorkflowOptionsMatchOwnershipPolicy(t *testing.T) {
	file, err := os.Open("../../.github/product-owners.json")
	if err != nil {
		t.Fatalf("open repository product ownership policy: %v", err)
	}
	defer file.Close()

	policy, err := productownership.Load(file)
	if err != nil {
		t.Fatalf("load repository product ownership policy: %v", err)
	}
	workflow, err := os.ReadFile("../../.github/workflows/product-acceptance.yml")
	if err != nil {
		t.Fatalf("read product acceptance workflow: %v", err)
	}

	content := string(workflow)
	optionsStart := strings.Index(content, "        options:\n")
	permissionsStart := strings.Index(content, "\npermissions:\n")
	if optionsStart < 0 || permissionsStart < 0 || permissionsStart <= optionsStart {
		t.Fatal("cannot locate product acceptance workflow options")
	}

	var got []string
	for _, line := range strings.Split(content[optionsStart+len("        options:\n"):permissionsStart], "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			got = append(got, strings.TrimPrefix(line, "- "))
		}
	}
	want := make([]string, 0, len(policy.Products))
	for name := range policy.Products {
		want = append(want, name)
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("product acceptance options = %v, want %v", got, want)
	}
}
