package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestValidateSharedBandwidthConfig_WithPackageId(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
		"internet_type":              "bgp",
		"share_bandwidth_package_id": "bwpack-test",
		"charge_mode":                "share_bandwidth",
		"bandwidth":                  0,
	})

	err := validateSharedBandwidthConfig(d)
	if err != nil {
		t.Fatalf("expected no error with valid shared bandwidth config, got: %s", err)
	}
}

func TestValidateSharedBandwidthConfig_WithPackageId_InvalidChargeMode(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
		"internet_type":              "bgp",
		"share_bandwidth_package_id": "bwpack-test",
		"charge_mode":                "bandwidth",
		"bandwidth":                  0,
	})

	err := validateSharedBandwidthConfig(d)
	if err == nil {
		t.Fatal("expected error with invalid charge_mode, got nil")
	}
}

func TestValidateSharedBandwidthConfig_WithoutPackageId(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
		"internet_type": "bgp",
		"charge_mode":   "bandwidth",
		"bandwidth":     2,
	})

	err := validateSharedBandwidthConfig(d)
	if err != nil {
		t.Fatalf("expected no error with valid regular config, got: %s", err)
	}
}

func TestValidateSharedBandwidthConfig_WithoutPackageId_InvalidChargeMode(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
		"internet_type": "bgp",
		"charge_mode":   "share_bandwidth",
		"bandwidth":     2,
	})

	err := validateSharedBandwidthConfig(d)
	if err == nil {
		t.Fatal("expected error with share_bandwidth mode without package_id, got nil")
	}
}

func TestValidateSharedBandwidthConfig_WithoutPackageId_ZeroBandwidth(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
		"internet_type": "bgp",
		"charge_mode":   "bandwidth",
		"bandwidth":     0,
	})

	err := validateSharedBandwidthConfig(d)
	if err == nil {
		t.Fatal("expected error with zero bandwidth without package_id, got nil")
	}
}
