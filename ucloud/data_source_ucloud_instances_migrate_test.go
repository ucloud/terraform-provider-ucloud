package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestUCloudInstancesMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1_is_boot": {
			StateVersion: 0,
			Attributes: map[string]string{
				"instances.0.disk_set.0.is_boot": "Yes",
				"instances.1.disk_set.1.is_boot": "No",
			},
			Expected: map[string]string{
				"instances.0.disk_set.0.is_boot": "true",
				"instances.1.disk_set.1.is_boot": "false",
			},
		},

		"v0_1_disk_type": {
			StateVersion: 0,
			Attributes: map[string]string{
				"instances.0.disk_set.0.type": "CLOUD_SSD",
				"instances.1.disk_set.1.type": "LOCAL_NORMAL",
			},
			Expected: map[string]string{
				"instances.0.disk_set.0.type": "cloud_ssd",
				"instances.1.disk_set.1.type": "local_normal",
			},
		},

		"v0_1_auto_renew": {
			StateVersion: 0,
			Attributes: map[string]string{
				"instances.0.auto_renew": "Yes",
				"instances.1.auto_renew": "No",
			},
			Expected: map[string]string{
				"instances.0.auto_renew": "true",
				"instances.1.auto_renew": "false",
			},
		},

		"v0_1_memory": {
			StateVersion: 0,
			Attributes: map[string]string{
				"instances.0.memory": "2048",
				"instances.1.memory": "1024",
			},
			Expected: map[string]string{
				"instances.0.memory": "2",
				"instances.1.memory": "1",
			},
		},

		"v0_1_charge_type": {
			StateVersion: 0,
			Attributes: map[string]string{
				"instances.0.charge_type": "Month",
				"instances.1.charge_type": "Dynamic",
			},
			Expected: map[string]string{
				"instances.0.charge_type": "month",
				"instances.1.charge_type": "dynamic",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "foo",
			Attributes: tc.Attributes,
		}

		is, err := dataSourceUCloudInstancesMigrateState(tc.StateVersion, is, tc.Meta)
		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}
