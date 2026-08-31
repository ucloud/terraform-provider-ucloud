// Package acceptancetest provides the shared, test-only provider harness used
// by independently maintained product acceptance suites.
package acceptancetest

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	providerucloud "github.com/terraform-providers/terraform-provider-ucloud/ucloud"
)

// Harness owns one provider instance so resource checks and product SDK
// clients observe the same configured runtime.
type Harness struct {
	Provider  *schema.Provider
	Providers map[string]terraform.ResourceProvider
}

// New creates an isolated provider harness for one product package.
func New() *Harness {
	provider := providerucloud.Provider().(*schema.Provider)
	return &Harness{
		Provider: provider,
		Providers: map[string]terraform.ResourceProvider{
			"ucloud": provider,
		},
	}
}

// PreCheck validates the common environment required by cloud acceptance
// tests without logging credential values.
func (h *Harness) PreCheck(t testing.TB) {
	t.Helper()
	for _, name := range []string{"UCLOUD_PUBLIC_KEY", "UCLOUD_PRIVATE_KEY", "UCLOUD_PROJECT_ID"} {
		if os.Getenv(name) == "" {
			t.Fatalf("%s must be set for acceptance tests", name)
		}
	}
	if os.Getenv("UCLOUD_REGION") == "" {
		log.Print("[INFO] Test: Using cn-bj2 as test region")
		if err := os.Setenv("UCLOUD_REGION", "cn-bj2"); err != nil {
			t.Fatalf("set UCLOUD_REGION: %v", err)
		}
	}
}

// CheckIDExists verifies that Terraform recorded a non-empty state ID.
func CheckIDExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("cannot find resource or data source %q", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("resource or data source %q has no ID", name)
		}
		return nil
	}
}

// CheckFileExists verifies an acceptance-test output file without reading it.
func CheckFileExists(path string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("stat acceptance output %q: %w", path, err)
		}
		return nil
	}
}

// ProductClient returns a test-only cached client for a product. The cache key
// is isolated from the production product client, which may be a package-local
// composite type. Call it after Terraform has configured the provider.
func (h *Harness) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	if h == nil || h.Provider == nil {
		return nil, fmt.Errorf("acceptance harness provider is nil")
	}
	runtime, ok := h.Provider.Meta().(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", h.Provider.Meta())
	}
	return runtime.ProductClient("acceptance:"+name, constructor)
}
