package product

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type testProduct struct {
	registration Registration
}

func (p testProduct) Registration() Registration {
	return p.registration
}

func testBinding(name string, registration Registration) Binding {
	return Bind(name, testProduct{registration: registration})
}

func TestRegister(t *testing.T) {
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"ucloud_existing": {Read: func(*schema.ResourceData, interface{}) error { return nil }},
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}

	err := Register(provider, testBinding("example", Registration{
		Name: "example",
		Resources: map[string]*schema.Resource{
			"ucloud_example": {Read: func(*schema.ResourceData, interface{}) error { return nil }},
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_example_items": {Read: func(*schema.ResourceData, interface{}) error { return nil }},
		},
	}))
	if err != nil {
		t.Fatalf("register product: %v", err)
	}

	if provider.ResourcesMap["ucloud_existing"] == nil {
		t.Fatal("existing resource was removed")
	}
	if provider.ResourcesMap["ucloud_example"] == nil {
		t.Fatal("product resource was not registered")
	}
	if provider.DataSourcesMap["ucloud_example_items"] == nil {
		t.Fatal("product data source was not registered")
	}
}

func TestRegisterRejectsInvalidProductsWithoutMutatingProvider(t *testing.T) {
	tests := map[string]struct {
		products []Binding
		want     string
	}{
		"empty name": {
			products: []Binding{testBinding("", Registration{
				Resources: map[string]*schema.Resource{"ucloud_example": {}},
			})},
			want: "name",
		},
		"empty registration": {
			products: []Binding{testBinding("example", Registration{Name: "example"})},
			want:     "at least one",
		},
		"invalid terraform name": {
			products: []Binding{testBinding("example", Registration{
				Name:      "example",
				Resources: map[string]*schema.Resource{"example": {}},
			})},
			want: "ucloud_",
		},
		"outside product namespace": {
			products: []Binding{testBinding("us3", Registration{
				Name:      "us3",
				Resources: map[string]*schema.Resource{"ucloud_uhost_instance": {}},
			})},
			want: "namespace",
		},
		"invalid legacy namespace": {
			products: []Binding{Bind("udisk", testProduct{registration: Registration{
				Name:      "udisk",
				Resources: map[string]*schema.Resource{"ucloud_disk": {}},
			}}, WithTerraformNamespaces("ucloud-disk"))},
			want: "Terraform namespace",
		},
		"nil resource": {
			products: []Binding{testBinding("example", Registration{
				Name:      "example",
				Resources: map[string]*schema.Resource{"ucloud_example": nil},
			})},
			want: "nil",
		},
		"duplicate existing resource": {
			products: []Binding{testBinding("existing", Registration{
				Name:      "existing",
				Resources: map[string]*schema.Resource{"ucloud_existing": {}},
			})},
			want: "already registered",
		},
		"duplicate product name": {
			products: []Binding{
				testBinding("example", Registration{
					Name:      "example",
					Resources: map[string]*schema.Resource{"ucloud_example_one": {}},
				}),
				testBinding("example", Registration{
					Name:      "example",
					Resources: map[string]*schema.Resource{"ucloud_example_two": {}},
				}),
			},
			want: "registered more than once",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			provider := &schema.Provider{
				ResourcesMap: map[string]*schema.Resource{"ucloud_existing": {}},
			}

			err := Register(provider, tc.products...)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Register() error = %v, want containing %q", err, tc.want)
			}
			if len(provider.ResourcesMap) != 1 || provider.ResourcesMap["ucloud_existing"] == nil {
				t.Fatalf("provider mutated after failed registration: %#v", provider.ResourcesMap)
			}
			if len(provider.DataSourcesMap) != 0 {
				t.Fatalf("data sources mutated after failed registration: %#v", provider.DataSourcesMap)
			}
		})
	}
}

func TestRegisterAllowsCoreOwnedLegacyTerraformNamespaces(t *testing.T) {
	provider := &schema.Provider{}
	err := Register(provider, Bind("udisk", testProduct{registration: Registration{
		Name: "udisk",
		Resources: map[string]*schema.Resource{
			"ucloud_disk":          {},
			"ucloud_disk_snapshot": {},
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_disks": {},
		},
	}}, WithTerraformNamespaces("disk", "disks")))
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if provider.ResourcesMap["ucloud_disk"] == nil || provider.DataSourcesMap["ucloud_disks"] == nil {
		t.Fatalf("legacy Terraform registrations missing: resources=%#v data_sources=%#v", provider.ResourcesMap, provider.DataSourcesMap)
	}
}

func TestRegisterRejectsProductNameThatDoesNotMatchCoreBinding(t *testing.T) {
	provider := &schema.Provider{}
	adapter := testProduct{registration: Registration{
		Name:      "uhost",
		Resources: map[string]*schema.Resource{"ucloud_uhost_instance": {}},
	}}

	err := Register(provider, Bind("us3", adapter))
	if err == nil || !strings.Contains(err.Error(), "bound") {
		t.Fatalf("Register() error = %v, want core binding error", err)
	}
}
