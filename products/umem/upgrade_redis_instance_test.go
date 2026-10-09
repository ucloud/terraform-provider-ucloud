package umem

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TestRedisInstanceUpgradeFromPublishedState covers the state/upgrade part of
// spec 7.3 for the vpc_id/subnet_id change: a state written by the last
// published provider (test-fixtures/redis_instance-state-v0.json, captured
// before the two attributes became required) must still be readable, and a
// configuration equivalent to that state must plan no unexpected diff.
func TestRedisInstanceUpgradeFromPublishedState(t *testing.T) {
	raw, err := os.ReadFile("test-fixtures/redis_instance-state-v0.json")
	if err != nil {
		t.Fatalf("read published state fixture: %v", err)
	}

	var state terraform.InstanceState
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatalf("decode published state fixture: %v", err)
	}

	r := resourceUCloudRedisInstance()

	// The current schema must still consume the published state without
	// dropping the attributes the user already has on disk.
	normalized := r.Data(&state).State()
	if normalized == nil {
		t.Fatal("current redis schema dropped the published state")
	}
	for _, attr := range []string{
		"id", "availability_zone", "standby_zone", "name", "tag", "instance_type",
		"engine_version", "password", "vpc_id", "subnet_id", "status",
	} {
		want := state.Attributes[attr]
		if want == "" {
			t.Fatalf("fixture is missing attribute %q", attr)
		}
		if got := normalized.Attributes[attr]; got != want {
			t.Errorf("published state attribute %q = %q, want %q", attr, got, want)
		}
	}

	// A configuration equivalent to the published state (only adding the two
	// attributes that are now required) must plan no change at all.
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"availability_zone": state.Attributes["availability_zone"],
		"standby_zone":      state.Attributes["standby_zone"],
		"name":              state.Attributes["name"],
		"tag":               state.Attributes["tag"],
		"instance_type":     state.Attributes["instance_type"],
		"engine_version":    state.Attributes["engine_version"],
		"password":          state.Attributes["password"],
		"vpc_id":            state.Attributes["vpc_id"],
		"subnet_id":         state.Attributes["subnet_id"],
		"charge_type":       state.Attributes["charge_type"],
		"duration":          1,
		"auto_backup":       state.Attributes["auto_backup"],
		"backup_begin_time": 3,
	})

	diff, err := r.Diff(&state, config, nil)
	if err != nil {
		t.Fatalf("plan against published state failed: %v", err)
	}
	var unexpected []string
	if diff != nil && !diff.Destroy {
		for name, attr := range diff.Attributes {
			if attr == nil {
				continue
			}
			// Computed attributes introduced after the published release
			// (block_set, proxy_set, read_mode) are "known after apply" on the
			// first plan against an old state; the first refresh fills them in
			// and no user value changes. Tolerate exactly that shape.
			if attr.NewComputed && attr.Old == "" && attr.New == "" {
				continue
			}
			unexpected = append(unexpected, name)
		}
	}
	if len(unexpected) > 0 {
		t.Errorf("published state produced an unexpected diff for %v: %#v", unexpected, diff.Attributes)
	}

	// The breaking part of the change is intentional: a configuration that still
	// omits the now required attributes no longer validates. Terraform core
	// reports this, so assert the schema fact the provider owns instead.
	for _, attr := range []string{"vpc_id", "subnet_id"} {
		field, ok := r.Schema[attr]
		if !ok {
			t.Fatalf("redis schema is missing %q", attr)
		}
		if !field.Required || field.Optional || field.Computed {
			t.Errorf("redis schema %q must be Required-only, got required=%t optional=%t computed=%t",
				attr, field.Required, field.Optional, field.Computed)
		}
		if !field.ForceNew {
			t.Errorf("redis schema %q must be ForceNew", attr)
		}
	}
}
