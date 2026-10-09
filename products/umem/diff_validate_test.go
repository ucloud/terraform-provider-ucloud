package umem

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestDiffValidateRedisInstanceType covers the instance_type change validator
// directly (it is a plain customdiff.ValidateChange callback): destroy plans
// carry an empty new value and must be accepted, creates carry an empty old
// value, in-place size changes within one family are allowed, and switching
// between active-standby (master) and distributed is rejected in both
// directions.
func TestDiffValidateRedisInstanceType(t *testing.T) {
	// destroy plan: empty new value means there is nothing to validate
	if err := diffValidateRedisInstanceType("redis-master-1", "", nil); err != nil {
		t.Fatalf("destroy plan (empty new value) must be accepted, got: %v", err)
	}

	// create plan: empty old value accepts any valid type
	if err := diffValidateRedisInstanceType("", "redis-master-1", nil); err != nil {
		t.Fatalf("create plan (empty old value) must be accepted, got: %v", err)
	}

	// resizing within the same family is an in-place change
	if err := diffValidateRedisInstanceType("redis-master-1", "redis-master-2", nil); err != nil {
		t.Fatalf("size change within master must be accepted, got: %v", err)
	}
	if err := diffValidateRedisInstanceType("redis-distributed-16", "redis-distributed-32", nil); err != nil {
		t.Fatalf("size change within distributed must be accepted, got: %v", err)
	}

	// switching families is rejected in both directions
	err := diffValidateRedisInstanceType("redis-master-1", "redis-distributed-16", nil)
	if err == nil {
		t.Fatal("expected error when switching from master to distributed, got nil")
	}
	if !strings.Contains(err.Error(), "instance_type") {
		t.Fatalf("expected instance_type in error, got: %v", err)
	}

	if err := diffValidateRedisInstanceType("redis-distributed-16", "redis-master-1", nil); err == nil {
		t.Fatal("expected error when switching from distributed to master, got nil")
	}
}

// TestRedisInstanceEngineVersionDiffValidation covers the engine_version rules
// of diffValidateRedisInstanceTypeAndEngineVersion: required for
// active-standby redis, rejected for distributed redis.
func TestRedisInstanceEngineVersionDiffValidation(t *testing.T) {
	r := resourceUCloudRedisInstance()

	// active-standby redis without engine_version must be rejected
	masterNoVersion := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-master-1",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
	})
	_, err := r.Diff(&terraform.InstanceState{}, masterNoVersion, nil)
	if err == nil {
		t.Fatal("expected error for active-standby redis without engine_version, got nil")
	}
	if !strings.Contains(err.Error(), "engine_version") {
		t.Fatalf("expected engine_version validation error, got: %v", err)
	}

	// distributed redis with engine_version must be rejected
	distributedWithVersion := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-distributed-16",
		"engine_version":    "4.0",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
	})
	_, err = r.Diff(&terraform.InstanceState{}, distributedWithVersion, nil)
	if err == nil {
		t.Fatal("expected error for distributed redis with engine_version, got nil")
	}
	if !strings.Contains(err.Error(), "engine_version") {
		t.Fatalf("expected engine_version validation error, got: %v", err)
	}

	// distributed redis without engine_version is valid
	distributedNoVersion := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-distributed-16",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
	})
	if _, err := r.Diff(&terraform.InstanceState{}, distributedNoVersion, nil); err != nil {
		t.Fatalf("distributed redis without engine_version should be valid, got: %v", err)
	}
}

// TestRedisInstanceBackupDiffValidation covers the two rejection branches of
// diffValidateBackup: distributed redis does not support backup settings at
// all, and backup_begin_time requires auto_backup on active-standby redis.
func TestRedisInstanceBackupDiffValidation(t *testing.T) {
	r := resourceUCloudRedisInstance()

	// distributed redis with backup settings must be rejected
	distributedBackup := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-distributed-16",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"auto_backup":       "enable",
		"backup_begin_time": 5,
	})
	_, err := r.Diff(&terraform.InstanceState{}, distributedBackup, nil)
	if err == nil {
		t.Fatal("expected error for distributed redis with backup settings, got nil")
	}
	if !strings.Contains(err.Error(), "not support backup") {
		t.Fatalf("expected distributed backup rejection error, got: %v", err)
	}

	// backup_begin_time without auto_backup must be rejected
	beginTimeOnly := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-master-1",
		"engine_version":    "4.0",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"backup_begin_time": 5,
	})
	_, err = r.Diff(&terraform.InstanceState{}, beginTimeOnly, nil)
	if err == nil {
		t.Fatal("expected error for backup_begin_time without auto_backup, got nil")
	}
	if !strings.Contains(err.Error(), "auto_backup") {
		t.Fatalf("expected auto_backup required error, got: %v", err)
	}

	// auto_backup with backup_begin_time on active-standby redis is valid
	fullBackup := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-master-1",
		"engine_version":    "4.0",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"auto_backup":       "enable",
		"backup_begin_time": 5,
	})
	if _, err := r.Diff(&terraform.InstanceState{}, fullBackup, nil); err != nil {
		t.Fatalf("active-standby redis with a full backup strategy should be valid, got: %v", err)
	}
}

// TestRedisInstanceRestartTriggerDiffValidation covers diffValidateRedisRestart:
// restart_trigger is an active-standby-only knob (the SDK exposes no restart
// API for distributed redis).
func TestRedisInstanceRestartTriggerDiffValidation(t *testing.T) {
	r := resourceUCloudRedisInstance()

	// distributed redis with restart_trigger must be rejected
	distributedRestart := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-distributed-16",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"restart_trigger":   "tf-acc-restart-1",
	})
	_, err := r.Diff(&terraform.InstanceState{}, distributedRestart, nil)
	if err == nil {
		t.Fatal("expected error for distributed redis with restart_trigger, got nil")
	}
	if !strings.Contains(err.Error(), "restart_trigger") {
		t.Fatalf("expected restart_trigger rejection error, got: %v", err)
	}

	// active-standby redis with restart_trigger is valid
	masterRestart := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"instance_type":     "redis-master-1",
		"engine_version":    "4.0",
		"name":              "tf-acc-redis",
		"vpc_id":            "ucloud-vpc-fake",
		"subnet_id":         "ucloud-subnet-fake",
		"restart_trigger":   "tf-acc-restart-1",
	})
	if _, err := r.Diff(&terraform.InstanceState{}, masterRestart, nil); err != nil {
		t.Fatalf("active-standby redis with restart_trigger should be valid, got: %v", err)
	}
}
