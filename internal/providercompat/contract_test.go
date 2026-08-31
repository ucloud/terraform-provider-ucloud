package providercompat

import (
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zclconf/go-cty/cty"
)

func testRead(*schema.ResourceData, interface{}) error   { return nil }
func testCreate(*schema.ResourceData, interface{}) error { return nil }
func testDelete(*schema.ResourceData, interface{}) error { return nil }
func testImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
func testMigrate(_ int, state *terraform.InstanceState, _ interface{}) (*terraform.InstanceState, error) {
	return state, nil
}
func testUpgrade(state map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	return state, nil
}

func TestMarshalCapturesProviderCompatibilityContract(t *testing.T) {
	timeout := 5 * time.Minute
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"region": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Example region",
				ValidateFunc: func(interface{}, string) ([]string, []error) { return nil, nil },
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ucloud_example": {
				Create:        testCreate,
				Read:          testRead,
				Delete:        testDelete,
				SchemaVersion: 1,
				MigrateState:  testMigrate,
				StateUpgraders: []schema.StateUpgrader{{
					Version: 0,
					Type:    cty.Object(map[string]cty.Type{"name": cty.String}),
					Upgrade: testUpgrade,
				}},
				Importer: &schema.ResourceImporter{State: testImport},
				Timeouts: &schema.ResourceTimeout{Create: &timeout},
				Schema: map[string]*schema.Schema{
					"name": {
						Type:          schema.TypeString,
						Optional:      true,
						ForceNew:      true,
						Default:       "example",
						ConflictsWith: []string{"other", "alias"},
					},
					"items": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"id": {Type: schema.TypeString, Computed: true},
						}},
					},
				},
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ucloud_examples": {
				Read: testRead,
				Schema: map[string]*schema.Schema{
					"ids": {Type: schema.TypeSet, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}, Set: schema.HashString},
				},
			},
		},
	}

	got, err := Marshal(provider)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	for _, want := range []string{
		"\"format_version\": 1",
		"\"ucloud_example\"",
		"\"ucloud_examples\"",
		"\"force_new\": true",
		"\"description\": \"Example region\"",
		"\"default\": \"\\\"example\\\"\"",
		"\"create\": \"testCreate\"",
		"\"importer\": \"testImport\"",
		"\"migrate_state\": \"testMigrate\"",
		"\"state_upgraders\"",
		"\"create\": \"5m0s\"",
		"\"set\": \"HashString\"",
		"\"resource\"",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("contract does not contain %q:\n%s", want, got)
		}
	}
}

func TestMarshalIsDeterministic(t *testing.T) {
	provider := &schema.Provider{ResourcesMap: map[string]*schema.Resource{
		"ucloud_z": {Read: testRead, Schema: map[string]*schema.Schema{"z": {Type: schema.TypeString, Computed: true}}},
		"ucloud_a": {Read: testRead, Schema: map[string]*schema.Schema{"a": {Type: schema.TypeString, Computed: true}}},
	}}

	first, err := Marshal(provider)
	if err != nil {
		t.Fatalf("first Marshal() error = %v", err)
	}
	second, err := Marshal(provider)
	if err != nil {
		t.Fatalf("second Marshal() error = %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("Marshal() is not deterministic:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestNormalizeFunctionNameIgnoresInitOrder(t *testing.T) {
	tests := map[string]string{
		"github.com/example/provider/ucloud.init.StringMatch.func17":                     "StringMatch",
		"github.com/example/provider/products/us3.init.StringMatch.func1":                "StringMatch",
		"github.com/example/provider/ucloud.resourceUCloudUS3Bucket.All.func2":           "All",
		"github.com/example/provider/ucloud.resourceUCloudUS3Bucket.StringInSlice.func1": "StringInSlice",
		"github.com/hashicorp/terraform-plugin-sdk/helper/schema.EnvDefaultFunc.func1":   "EnvDefaultFunc",
		"github.com/example/provider/ucloud.resourceUCloudUS3BucketCreate":               "resourceUCloudUS3BucketCreate",
	}
	for name, want := range tests {
		if got := normalizeFunctionName(name); got != want {
			t.Fatalf("normalizeFunctionName(%q) = %q, want %q", name, got, want)
		}
	}
}
