package uk8s

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestClusterReleasedStateUpgradePlan(t *testing.T) {
	for _, legacy := range []string{"n-basic-2", "n-highcpu-1", "n-customized-2-6", "o-basic-2", "N-basic-2"} {
		for _, format := range []string{"flatmap", "json"} {
			t.Run(legacy+"/"+format, func(t *testing.T) {
				old := testUK8SClusterReleasedSchema()
				current := resourceUCloudUK8SCluster()
				if !old.CoreConfigSchema().ImpliedType().Equals(current.StateUpgraders[0].Type) {
					t.Fatal("upgrader type does not match the published schema")
				}
				config := masterTestConfig(map[string]interface{}{
					"instance_type": legacy, "boot_disk_type": "cloud_rssd",
				})
				raw := terraform.NewResourceConfigRaw(config)
				for _, r := range []*schema.Resource{old, current} {
					if _, errs := r.Validate(raw); len(errs) > 0 {
						t.Fatal(errs)
					}
				}
				// Apply with the released schema, substituting only the cloud call.
				old.Create = func(d *schema.ResourceData, _ interface{}) error {
					d.SetId("uk8s-existing")
					return nil
				}
				plan, err := old.Diff(nil, raw, nil)
				if err != nil {
					t.Fatal(err)
				}
				state, err := old.Apply(nil, plan, nil)
				if err != nil {
					t.Fatal(err)
				}
				// The cluster Read preserves the creation-only master block.
				current.Read = func(_ *schema.ResourceData, _ interface{}) error { return nil }
				if format == "flatmap" {
					state, err = current.Refresh(state, nil)
				} else {
					value, valueErr := schema.StateValueFromInstanceState(state, old.CoreConfigSchema().ImpliedType())
					if valueErr != nil {
						t.Fatal(valueErr)
					}
					jsonState, valueErr := schema.StateValueToJSONMap(value, old.CoreConfigSchema().ImpliedType())
					if valueErr != nil {
						t.Fatal(valueErr)
					}
					upgraded, upgradeErr := current.StateUpgraders[0].Upgrade(jsonState, nil)
					if upgradeErr != nil {
						t.Fatal(upgradeErr)
					}
					value, err = schema.JSONMapToStateValue(upgraded, current.CoreConfigSchema())
					if err == nil {
						state, err = current.ShimInstanceStateFromValue(value)
					}
				}
				if err != nil {
					t.Fatal(err)
				}
				if state.ID != "uk8s-existing" || state.Attributes["master.0.instance_type"] != legacy {
					t.Fatalf("upgrade changed the ID or legacy specification: %#v", state)
				}
				plan, err = current.Diff(state, raw, nil)
				if err != nil {
					t.Fatal(err)
				}
				if plan != nil && !plan.Empty() {
					t.Fatalf("upgrade produced an unexpected plan: %#v", plan)
				}
			})
		}
	}
}

func TestClusterUpgradePreservesHistoricalValues(t *testing.T) {
	for _, legacy := range []string{"", " N-basic-2 ", "unexpected-historical-value"} {
		t.Run(legacy, func(t *testing.T) {
			raw := map[string]interface{}{
				"id":     " historical-cluster-id ",
				"master": []interface{}{map[string]interface{}{"instance_type": legacy}},
			}
			before, err := json.Marshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			upgraded, err := resourceUCloudUK8SCluster().StateUpgraders[0].Upgrade(raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			after, err := json.Marshal(upgraded)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatalf("historical state changed: %s -> %s, error: %v", before, after, err)
			}
		})
	}
}
