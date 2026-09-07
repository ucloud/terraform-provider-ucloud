package uhost

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestValidateInstanceLoginModeValues(t *testing.T) {
	tests := []struct {
		name            string
		loginMode       string
		keyPairID       string
		rootPassword    string
		hasRootPassword bool
		wantErr         bool
	}{
		{name: "default password login without key pair", rootPassword: "wA1234567", hasRootPassword: true},
		{name: "explicit password login without key pair", loginMode: "Password", rootPassword: "wA1234567", hasRootPassword: true},
		{name: "key pair login with key pair id", loginMode: "KeyPair", keyPairID: "keypair-test"},
		{name: "key pair id requires login mode", keyPairID: "keypair-test", wantErr: true},
		{name: "password login rejects key pair id", loginMode: "Password", keyPairID: "keypair-test", wantErr: true},
		{name: "key pair login requires key pair id", loginMode: "KeyPair", wantErr: true},
		{
			name:            "key pair login rejects root password",
			loginMode:       "KeyPair",
			keyPairID:       "keypair-test",
			rootPassword:    "wA1234567",
			hasRootPassword: true,
			wantErr:         true,
		},
		{
			name:            "key pair login rejects empty root password",
			loginMode:       "KeyPair",
			keyPairID:       "keypair-test",
			hasRootPassword: true,
			wantErr:         true,
		},
		{name: "unsupported login mode", loginMode: "ImagePasswd", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateInstanceLoginModeValues(
				test.loginMode,
				test.keyPairID,
				test.rootPassword,
				test.hasRootPassword,
			)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateInstanceLoginModeValues() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateInstanceLoginModeIgnoresEmptyRootPasswordState(t *testing.T) {
	resource := &schema.Resource{
		CustomizeDiff: validateInstanceLoginMode,
		Schema: map[string]*schema.Schema{
			"login_mode": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_pair_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"root_password": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
	config := map[string]interface{}{
		"login_mode":  "KeyPair",
		"key_pair_id": "keypair-test",
	}
	state := &terraform.InstanceState{
		ID: "uhost-test",
		Attributes: map[string]string{
			"login_mode":    "KeyPair",
			"key_pair_id":   "keypair-test",
			"root_password": "",
		},
	}

	if _, err := resource.Diff(state, terraform.NewResourceConfigRaw(config), nil); err != nil {
		t.Fatalf("Diff() with empty root_password state returned error: %s", err)
	}

	config["root_password"] = "wA1234567"
	if _, err := resource.Diff(nil, terraform.NewResourceConfigRaw(config), nil); err == nil {
		t.Fatal("Diff() with key pair login and configured root_password returned no error")
	}
}
