package providercompat

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
)

func TestBaselinesPreserveFullContract(test *testing.T) {
	provider, metadata := baselineFixture()
	baselines, err := Baselines(provider, metadata)
	if err != nil {
		test.Fatal(err)
	}
	if len(baselines) != len(metadata)+1 {
		test.Fatalf("baseline count = %d, want %d", len(baselines), len(metadata)+1)
	}

	var combined Contract
	if err := json.Unmarshal(baselines["core"], &combined); err != nil {
		test.Fatal(err)
	}
	if len(combined.Resources) != 0 || len(combined.DataSources) != 0 {
		test.Fatal("core baseline contains product registrations")
	}
	combined.Resources = make(map[string]ResourceContract)
	combined.DataSources = make(map[string]ResourceContract)
	for _, product := range metadata {
		var partition Contract
		if err := json.Unmarshal(baselines[product.Name], &partition); err != nil {
			test.Fatal(err)
		}
		if len(partition.Provider) != 0 || partition.Configure != "" {
			test.Fatalf("product %s baseline contains provider configuration", product.Name)
		}
		if len(partition.Resources) != len(product.ResourceTypes) || len(partition.DataSources) != len(product.DataSourceTypes) {
			test.Fatalf("product %s has unexpected registrations: %#v", product.Name, partition)
		}
		if partition.FormatVersion != combined.FormatVersion {
			test.Fatalf("product %s has inconsistent format version", product.Name)
		}
		for name, resource := range partition.Resources {
			if _, exists := combined.Resources[name]; exists {
				test.Fatalf("resource %s is in multiple baselines", name)
			}
			combined.Resources[name] = resource
		}
		for name, dataSource := range partition.DataSources {
			if _, exists := combined.DataSources[name]; exists {
				test.Fatalf("data source %s is in multiple baselines", name)
			}
			combined.DataSources[name] = dataSource
		}
	}
	reconstructed, err := marshalContract(combined)
	if err != nil {
		test.Fatal(err)
	}
	original, err := Marshal(provider)
	if err != nil {
		test.Fatal(err)
	}
	if !bytes.Equal(reconstructed, original) {
		test.Fatalf("partitioning changed the contract:\noriginal: %s\nreconstructed: %s", original, reconstructed)
	}
	repeated, err := Baselines(provider, metadata)
	if err != nil {
		test.Fatal(err)
	}
	if !reflect.DeepEqual(baselines, repeated) {
		test.Fatal("baseline generation is not deterministic or mutated the provider")
	}
}

func TestBaselinesRejectIncompleteOrAmbiguousOwnership(test *testing.T) {
	for _, scenario := range []struct {
		name   string
		change func(*schema.Provider, *[]productcatalog.OwnershipMetadata)
		want   string
	}{
		{
			name: "unowned resource",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				provider.ResourcesMap["ucloud_unowned"] = &schema.Resource{}
			},
			want: "resources have no product owner: ucloud_unowned",
		},
		{
			name: "unowned data source",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				provider.DataSourcesMap["ucloud_unowned"] = &schema.Resource{}
			},
			want: "data sources have no product owner: ucloud_unowned",
		},
		{
			name: "missing resource",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[0].ResourceTypes = append((*metadata)[0].ResourceTypes, "ucloud_missing")
			},
			want: `"ucloud_missing" is not registered`,
		},
		{
			name: "missing data source",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[2].DataSourceTypes = append((*metadata)[2].DataSourceTypes, "ucloud_missing")
			},
			want: `"ucloud_missing" is not registered`,
		},
		{
			name: "resource assigned twice",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[1].ResourceTypes = append((*metadata)[1].ResourceTypes, "ucloud_instance")
			},
			want: `"ucloud_instance" is not registered or is assigned more than once`,
		},
		{
			name: "data source assigned twice",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[2].DataSourceTypes = append((*metadata)[2].DataSourceTypes, "ucloud_projects")
			},
			want: `"ucloud_projects" is not registered or is assigned more than once`,
		},
		{
			name: "duplicate product",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				*metadata = append(*metadata, (*metadata)[0])
			},
			want: `duplicate baseline product "compute"`,
		},
		{
			name: "reserved core name",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[0].Name = "core"
			},
			want: `invalid baseline product name "core"`,
		},
		{
			name: "reserved all name",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[0].Name = "all"
			},
			want: `invalid baseline product name "all"`,
		},
		{
			name: "product path traversal",
			change: func(provider *schema.Provider, metadata *[]productcatalog.OwnershipMetadata) {
				(*metadata)[0].Name = "../outside"
			},
			want: `invalid baseline product name "../outside"`,
		},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			provider, metadata := baselineFixture()
			scenario.change(provider, &metadata)
			baselines, err := Baselines(provider, metadata)
			if err == nil || !strings.Contains(err.Error(), scenario.want) {
				test.Fatalf("Baselines() error = %v, want %q", err, scenario.want)
			}
			if baselines != nil {
				test.Fatal("failed generation returned partial baselines")
			}
		})
	}
}

func baselineFixture() (*schema.Provider, []productcatalog.OwnershipMetadata) {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"region": {Type: schema.TypeString, Required: true},
		},
		ConfigureFunc: func(*schema.ResourceData) (interface{}, error) { return nil, nil },
		ResourcesMap: map[string]*schema.Resource{
			"ucloud_instance": {
				Read: testRead,
				Schema: map[string]*schema.Schema{
					"name": {Type: schema.TypeString, Optional: true, Default: "example"},
				},
			},
			"ucloud_vpc": {Read: testRead},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ucloud_projects": {Read: testRead},
		},
	}
	return provider, []productcatalog.OwnershipMetadata{
		{Name: "compute", ResourceTypes: []string{"ucloud_instance"}},
		{Name: "network", ResourceTypes: []string{"ucloud_vpc"}},
		{Name: "account", DataSourceTypes: []string{"ucloud_projects"}},
	}
}
