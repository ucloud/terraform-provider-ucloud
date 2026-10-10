package umem

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestRedisInstanceDestroyDiffSkipsValidations reproduces the failure seen in
// acceptance tests: after a failed apply, the test framework destroys the
// dangling resource, and the destroy plan (empty config) ran CustomizeDiff
// validations that only make sense for create/update, producing a spurious
// "the argument \"auto_backup\" is required when set \"backup_begin_time\"".
func TestRedisInstanceDestroyDiffSkipsValidations(t *testing.T) {
	r := resourceUCloudRedisInstance()

	state := &terraform.InstanceState{
		ID: "uredis-fake-id",
		Attributes: map[string]string{
			"id":                "uredis-fake-id",
			"availability_zone": "cn-wlcb-01",
			"name":              "tf-acc-redis-renamed",
			"instance_type":     "redis-master-2",
			"engine_version":    "4.0",
			"backup_begin_time": "3",
			"auto_backup":       "disable",
			"port":              "6379",
			"tag":               "tf-acc",
			"charge_type":       "month",
			"duration":          "1",
		},
	}

	// A destroy plan diffs the prior state against an empty config.
	emptyConfig := terraform.NewResourceConfigRaw(map[string]interface{}{})

	diff, err := r.Diff(state, emptyConfig, nil)
	if err != nil {
		t.Fatalf("destroy diff must not fail validation, got: %v", err)
	}
	if diff != nil && !diff.Destroy && !diff.Empty() {
		// Not fatal for this regression test; just record what we got.
		t.Logf("destroy diff: %#v", diff)
	}
}

// TestRedisInstanceCreateDiffStillValidates guards against over-suppression:
// the destroy short-circuit must not disable validation for real create plans.
func TestRedisInstanceCreateDiffStillValidates(t *testing.T) {
	r := resourceUCloudRedisInstance()

	// standby_zone equal to availability_zone must still be rejected.
	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-wlcb-01",
		"standby_zone":      "cn-wlcb-01",
		"instance_type":     "redis-master-1",
		"engine_version":    "4.0",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
	})

	_, err := r.Diff(&terraform.InstanceState{}, cfg, nil)
	if err == nil {
		t.Fatal("expected validation error for identical standby_zone, got nil")
	}
	if !strings.Contains(err.Error(), "standby_zone") {
		t.Fatalf("expected standby_zone validation error, got: %v", err)
	}
}

// TestRedisInstanceBlockCntValidation covers the block_cnt rule: it is only
// valid for distributed redis and must be rejected for active-standby redis.
func TestRedisInstanceBlockCntValidation(t *testing.T) {
	r := resourceUCloudRedisInstance()

	// active-standby (master) redis must reject block_cnt.
	masterCfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"standby_zone":      "cn-sh2-02",
		"instance_type":     "redis-master-1",
		"engine_version":    "4.0",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"block_cnt":         4,
	})
	_, err := r.Diff(&terraform.InstanceState{}, masterCfg, nil)
	if err == nil {
		t.Fatal("expected block_cnt to be rejected for active-standby redis, got nil")
	}
	if !strings.Contains(err.Error(), "block_cnt") {
		t.Fatalf("expected block_cnt validation error, got: %v", err)
	}

	// distributed redis must accept block_cnt (and must not set engine_version).
	distributedCfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-distributed-16",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"block_cnt":         4,
	})
	if _, err := r.Diff(&terraform.InstanceState{}, distributedCfg, nil); err != nil {
		t.Fatalf("distributed redis with block_cnt should be valid, got: %v", err)
	}
}
