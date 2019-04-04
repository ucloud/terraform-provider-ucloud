package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestUCloudEipsMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1_charge_type": {
			StateVersion: 0,
			Attributes: map[string]string{
				"eips.0.charge_type": "Month",
				"eips.1.charge_type": "Dynamic",
			},
			Expected: map[string]string{
				"eips.0.charge_type": "month",
				"eips.1.charge_type": "dynamic",
			},
		},

		"v0_1_charge_mode": {
			StateVersion: 0,
			Attributes: map[string]string{
				"eips.0.charge_type": "Bandwidth",
				"eips.1.charge_type": "ShareBandwidth",
			},
			Expected: map[string]string{
				"eips.0.charge_type": "bandwidth",
				"eips.1.charge_type": "share_bandwidth",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "foo",
			Attributes: tc.Attributes,
		}

		is, err := dataSourceUCloudEipsMigrateState(tc.StateVersion, is, tc.Meta)
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
