package acceptancetest_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestNewBuildsAnIsolatedValidProvider(t *testing.T) {
	first := acceptancetest.New()
	second := acceptancetest.New()
	if first.Provider == second.Provider {
		t.Fatal("New() reused a provider instance")
	}
	if first.Providers["ucloud"] != first.Provider {
		t.Fatal("Providers does not expose the harness provider")
	}
	if err := first.Provider.InternalValidate(); err != nil {
		t.Fatalf("provider validation failed: %v", err)
	}
}

func TestProductClientRejectsUnconfiguredProvider(t *testing.T) {
	harness := acceptancetest.New()
	_, err := harness.ProductClient("example", func(
		*ucloud.Config,
		*auth.Credential,
		[]ucloud.HttpRequestHandler,
	) interface{} {
		return nil
	})
	if err == nil {
		t.Fatal("ProductClient() accepted an unconfigured provider")
	}
}
