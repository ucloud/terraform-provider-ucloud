package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestUCloudDiskAttachmentMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     string
		Meta         interface{}
	}{
		"v0_1_id": {
			StateVersion: 0,
			ID:           "disk#disk-abcd:uhost#uhost-abcd",
			Attributes: map[string]string{
				"id": "disk#disk-abcd:uhost#uhost-abcd",
			},
			Expected: "disk-abcd:uhost-abcd",
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}

		is, err := resourceUCloudDiskAttachmentMigrateState(tc.StateVersion, is, tc.Meta)
		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.ID != tc.Expected {
			t.Fatalf("bad: %s\n\n expected: ID -> %#v\n got: ID -> %#v\n in: %#v",
				tn, tc.Expected, is.ID, is.Attributes)
		}

		if is.Attributes["id"] != tc.Expected {
			t.Fatalf("bad: %s\n\n expected: ID -> %#v\n got: ID -> %#v\n in: %#v",
				tn, tc.Expected, is.Attributes["id"], is.Attributes)
		}
	}
}
