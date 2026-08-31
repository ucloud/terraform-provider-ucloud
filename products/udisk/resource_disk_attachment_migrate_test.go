package udisk

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestDiskAttachmentMigrateStateV0(t *testing.T) {
	state := &terraform.InstanceState{
		ID: "disk#disk-abcd:uhost#uhost-abcd",
		Attributes: map[string]string{
			"id": "disk#disk-abcd:uhost#uhost-abcd",
		},
	}

	state, err := resourceUCloudDiskAttachmentMigrateState(0, state, nil)
	if err != nil {
		t.Fatalf("migrate attachment state: %v", err)
	}
	if state.ID != "disk-abcd:uhost-abcd" || state.Attributes["id"] != state.ID {
		t.Fatalf("migrated state = %#v, want disk-abcd:uhost-abcd", state)
	}
}

func TestDiskAttachmentMigrateStatePreservesLegacyWhitespaceIDs(t *testing.T) {
	state := &terraform.InstanceState{
		ID:         "disk# :uhost# ",
		Attributes: map[string]string{"id": "disk# :uhost# "},
	}

	state, err := resourceUCloudDiskAttachmentMigrateState(0, state, nil)
	if err != nil {
		t.Fatalf("migrate legacy whitespace attachment state: %v", err)
	}
	if state.ID != " : " || state.Attributes["id"] != state.ID {
		t.Fatalf("migrated state = %#v, want legacy whitespace ID preserved", state)
	}
}
