package vpc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// secGroupRuleUnknownValue is the marker the SDK substitutes for a value that
// is not known until apply, e.g. one derived from another resource. hcl2shim
// owns the marker but sits under the SDK's internal tree, so it cannot be
// imported here, and the literal has to be pinned down in this file instead.
const secGroupRuleUnknownValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

// secGroupRuleDiffConfig builds a configuration the rule resource accepts, with
// the two fields the port check reads left for the caller to set.
func secGroupRuleDiffConfig(protocolType, dstPort string) map[string]interface{} {
	config := map[string]interface{}{
		"sec_group_id":  "secgroup-abc123",
		"direction":     "Ingress",
		"protocol_type": protocolType,
		"ip_range":      "10.0.0.0/8",
		"rule_action":   "Accept",
		"priority":      50,
	}
	if dstPort != "" {
		config["dst_port"] = dstPort
	}
	return config
}

// TestDiffValidateSecGroupRulePort covers the plan-time check that a port is
// set exactly when the protocol carries one. The check runs through the
// resource rather than the function alone, so it also fails if the CustomizeDiff
// registration is dropped.
//
// Both mistakes it catches are otherwise silent: a TCP rule that reaches the API
// without a port can come back as a rule open on every port, and a port on an
// ICMP rule is normalized away on read, which leaves a diff that no apply can
// ever clear.
func TestDiffValidateSecGroupRulePort(t *testing.T) {
	tests := []struct {
		name         string
		protocolType string
		dstPort      string
		wantErr      bool
	}{
		{name: "tcp without a port", protocolType: "TCP", wantErr: true},
		{name: "udp without a port", protocolType: "UDP", wantErr: true},
		{name: "tcp with a port", protocolType: "TCP", dstPort: "22"},
		{name: "udp with a port", protocolType: "UDP", dstPort: "80,443"},
		{name: "tcp with a range", protocolType: "TCP", dstPort: "443,2000-10000"},
		{name: "icmp with a port", protocolType: "ICMP", dstPort: "22", wantErr: true},
		{name: "icmpv6 with a port", protocolType: "ICMPv6", dstPort: "22", wantErr: true},
		{name: "all with a port", protocolType: "ALL", dstPort: "22", wantErr: true},
		{name: "icmp without a port", protocolType: "ICMP"},
		{name: "icmpv6 without a port", protocolType: "ICMPv6"},
		{name: "all without a port", protocolType: "ALL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := secGroupRuleDiffConfig(test.protocolType, test.dstPort)

			_, err := resourceUCloudSecGroupRule().Diff(nil, terraform.NewResourceConfigRaw(config), nil)
			if (err != nil) != test.wantErr {
				t.Fatalf("Diff() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

// TestDiffValidateSecGroupRulePortSkipsUnknownValues covers the guard that
// leaves the check alone until both fields are known. Without it an unknown
// protocol_type reads back as the empty string, falls through to the
// port-dependent branch, and rejects a configuration that is perfectly good --
// which is exactly the mistake the check exists to prevent, aimed at the user
// instead of the API.
func TestDiffValidateSecGroupRulePortSkipsUnknownValues(t *testing.T) {
	// The marker is pinned down on remark, a field the check never reads, so
	// that this verdict cannot be confused with the check's own: if the SDK
	// ever changes the marker, the failure says the marker is stale rather than
	// leaving the case below to fail with a message about dst_port.
	markerConfig := secGroupRuleDiffConfig("TCP", "22")
	markerConfig["remark"] = secGroupRuleUnknownValue
	markerDiff, err := resourceUCloudSecGroupRule().Diff(nil, terraform.NewResourceConfigRaw(markerConfig), nil)
	if err != nil {
		t.Fatalf("Diff() with an unknown remark returned error: %s", err)
	}
	if attr := markerDiff.Attributes["remark"]; attr == nil || !attr.NewComputed {
		t.Fatalf("%q is no longer the marker the SDK uses for an unknown value", secGroupRuleUnknownValue)
	}

	config := secGroupRuleDiffConfig(secGroupRuleUnknownValue, "")

	if _, err := resourceUCloudSecGroupRule().Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
		t.Fatalf("Diff() with an unknown protocol_type returned error: %s", err)
	}
}
