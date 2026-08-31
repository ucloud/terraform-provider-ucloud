package unet

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestValidateSharedBandwidthConfig(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]interface{}
		wantError bool
	}{
		{
			name: "with package id",
			values: map[string]interface{}{
				"internet_type":              "bgp",
				"share_bandwidth_package_id": "bwpack-test",
				"charge_mode":                "share_bandwidth",
				"bandwidth":                  0,
			},
		},
		{
			name: "with package id and invalid charge mode",
			values: map[string]interface{}{
				"internet_type":              "bgp",
				"share_bandwidth_package_id": "bwpack-test",
				"charge_mode":                "bandwidth",
				"bandwidth":                  0,
			},
			wantError: true,
		},
		{
			name: "without package id",
			values: map[string]interface{}{
				"internet_type": "bgp",
				"charge_mode":   "bandwidth",
				"bandwidth":     2,
			},
		},
		{
			name: "without package id and invalid charge mode",
			values: map[string]interface{}{
				"internet_type": "bgp",
				"charge_mode":   "share_bandwidth",
				"bandwidth":     2,
			},
			wantError: true,
		},
		{
			name: "without package id and zero bandwidth",
			values: map[string]interface{}{
				"internet_type": "bgp",
				"charge_mode":   "bandwidth",
				"bandwidth":     0,
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, test.values)
			err := validateSharedBandwidthConfig(data)
			if (err != nil) != test.wantError {
				t.Fatalf("validateSharedBandwidthConfig() error = %v, want error = %t", err, test.wantError)
			}
		})
	}
}

func TestSecurityGroupRuleHash(t *testing.T) {
	rule := map[string]interface{}{
		"port_range": "80",
		"protocol":   "tcp",
		"cidr_block": "192.168.0.0/16",
		"policy":     "accept",
		"priority":   "high",
	}

	const want = 2629295509
	if got := resourceucloudSecurityGroupRuleHash(rule); got != want {
		t.Fatalf("resourceucloudSecurityGroupRuleHash() = %v, want %v", got, want)
	}
}

func TestLegacyEIPMigrationCases(t *testing.T) {
	tests := map[string]struct {
		attributes map[string]string
		want       map[string]string
	}{
		"charge type": {
			attributes: map[string]string{
				"eips.0.charge_type": "Month",
				"eips.1.charge_type": "Dynamic",
			},
			want: map[string]string{
				"eips.0.charge_type": "month",
				"eips.1.charge_type": "dynamic",
			},
		},
		"charge mode": {
			attributes: map[string]string{
				"eips.0.charge_mode": "Bandwidth",
				"eips.1.charge_mode": "ShareBandwidth",
			},
			want: map[string]string{
				"eips.0.charge_mode": "bandwidth",
				"eips.1.charge_mode": "share_bandwidth",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			state := &terraform.InstanceState{ID: "foo", Attributes: test.attributes}
			state, err := dataSourceUCloudEipsMigrateState(0, state, nil)
			if err != nil {
				t.Fatalf("migrate EIPs: %v", err)
			}
			for key, want := range test.want {
				if got := state.Attributes[key]; got != want {
					t.Errorf("attribute %q = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestLegacyEIPAssociationMigration(t *testing.T) {
	state := &terraform.InstanceState{
		ID: "eip#eip-abcd:instance#uhost-abcd",
		Attributes: map[string]string{
			"id": "eip#eip-abcd:instance#uhost-abcd",
		},
	}

	state, err := resourceUCloudEIPAssociationMigrateState(0, state, nil)
	if err != nil {
		t.Fatalf("migrate EIP association: %v", err)
	}
	if state.ID != "eip-abcd:uhost-abcd" || state.Attributes["id"] != state.ID {
		t.Fatalf("migrated association = %#v, want eip-abcd:uhost-abcd", state)
	}
}

func TestParseAssociationInfo(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want associationInfo
	}{
		{
			name: "minimal fields",
			id:   "a#b:c#d",
			want: associationInfo{PrimaryType: "a", PrimaryId: "b", ResourceType: "c", ResourceId: "d"},
		},
		{
			name: "normal resource id",
			id:   "eip#eip-abcd:instance#uhost-abcd",
			want: associationInfo{PrimaryType: "eip", PrimaryId: "eip-abcd", ResourceType: "instance", ResourceId: "uhost-abcd"},
		},
		{
			name: "ids with ordinary punctuation",
			id:   "network#primary.id_123:resource_type#resource-id_123",
			want: associationInfo{PrimaryType: "network", PrimaryId: "primary.id_123", ResourceType: "resource_type", ResourceId: "resource-id_123"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseAssociationInfo(test.id)
			if err != nil {
				t.Fatalf("parseAssociationInfo(%q): %v", test.id, err)
			}
			if *got != test.want {
				t.Errorf("parseAssociationInfo(%q) = %#v, want %#v", test.id, *got, test.want)
			}
		})
	}
}

func TestParseAssociationInfoRejectsMalformedAndEmptySegments(t *testing.T) {
	ids := []string{
		"",
		"no-delimiters",
		"eip#eip-abcd",
		"eip#eip-abcd:instance",
		"eip#eip-abcd:instance#",
		"eip#eip-abcd:#uhost-abcd",
		"eip#:instance#uhost-abcd",
		"eip$#eip-abcd:instance#uhost-abcd",
		"eip#eip-abcd:instance$#uhost-abcd",
	}

	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			if got, err := parseAssociationInfo(id); err == nil {
				t.Errorf("parseAssociationInfo(%q) = %#v, want error", id, got)
			}
		})
	}
}
