package ucloud

import "testing"

func TestValidateInstanceLoginModeValues(t *testing.T) {
	tests := []struct {
		name            string
		loginMode       string
		keyPairID       string
		rootPassword    string
		hasRootPassword bool
		wantErr         bool
	}{
		{
			name:            "default password login without key pair",
			rootPassword:    "wA1234567",
			hasRootPassword: true,
		},
		{
			name:            "explicit password login without key pair",
			loginMode:       "Password",
			rootPassword:    "wA1234567",
			hasRootPassword: true,
		},
		{
			name:      "key pair login with key pair id",
			loginMode: "KeyPair",
			keyPairID: "keypair-test",
		},
		{
			name:      "key pair id requires login mode",
			keyPairID: "keypair-test",
			wantErr:   true,
		},
		{
			name:      "password login rejects key pair id",
			loginMode: "Password",
			keyPairID: "keypair-test",
			wantErr:   true,
		},
		{
			name:      "key pair login requires key pair id",
			loginMode: "KeyPair",
			wantErr:   true,
		},
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
		{
			name:      "unsupported login mode",
			loginMode: "ImagePasswd",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstanceLoginModeValues(tt.loginMode, tt.keyPairID, tt.rootPassword, tt.hasRootPassword)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateInstanceLoginModeValues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
