package uhost

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestValidateInstanceUDSet(t *testing.T) {
	resource := &schema.Resource{
		CustomizeDiff: diffValidateInstanceUDSet,
		Schema: map[string]*schema.Schema{
			"udset_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"udhost_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"host_binding": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}

	tests := []struct {
		name    string
		newRes  bool
		config  map[string]interface{}
		state   map[string]string
		wantErr string
	}{
		{
			name:   "create without dedicated zone params",
			newRes: true,
			config: map[string]interface{}{},
		},
		{
			name:   "create in dedicated zone",
			newRes: true,
			config: map[string]interface{}{"udset_id": "udset-123"},
		},
		{
			name:   "create with udset_id and udhost_id",
			newRes: true,
			config: map[string]interface{}{"udset_id": "udset-123", "udhost_id": "udhost-456"},
		},
		{
			name:    "udhost_id requires udset_id",
			newRes:  true,
			config:  map[string]interface{}{"udhost_id": "udhost-456"},
			wantErr: `"udset_id" is required when set "udhost_id"`,
		},
		{
			name:    "host_binding requires udset_id",
			newRes:  true,
			config:  map[string]interface{}{"host_binding": true},
			wantErr: `"udset_id" is required when set "host_binding"`,
		},
		{
			name:   "host_binding false without udset_id is accepted",
			newRes: true,
			config: map[string]interface{}{"host_binding": false},
		},
		{
			name:    "changing udset_id is rejected",
			config:  map[string]interface{}{"udset_id": "udset-456"},
			state:   map[string]string{"udset_id": "udset-123"},
			wantErr: `the "udset_id" cannot be changed`,
		},
		{
			name:   "same udset_id is kept",
			config: map[string]interface{}{"udset_id": "udset-123"},
			state:  map[string]string{"udset_id": "udset-123"},
		},
		{
			name:    "adding udset_id to existing instance is rejected",
			config:  map[string]interface{}{"udset_id": "udset-123"},
			state:   map[string]string{},
			wantErr: `the "udset_id" cannot be changed`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var state *terraform.InstanceState
			if test.state != nil {
				state = &terraform.InstanceState{
					ID:         "uhost-test",
					Attributes: test.state,
				}
			}

			_, err := resource.Diff(state, terraform.NewResourceConfigRaw(test.config), nil)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("Diff() returned unexpected error: %s", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Diff() returned no error, want error containing %q", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Diff() error = %q, want containing %q", err.Error(), test.wantErr)
			}
		})
	}
}
