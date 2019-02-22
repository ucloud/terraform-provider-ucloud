package ucloud

import (
	"path/filepath"
	"testing"
)

const (
	testValueConfigPublicKey  = "tf-acc-public-key"
	testValueConfigPrivateKey = "tf-acc-private-key"
	testValueConfigProfile    = "default"
)

func TestConfigLoadCredential(t *testing.T) {
	config := &Config{
		Profile:               testValueConfigProfile,
		SharedCredentialsFile: filepath.Join(".", "test-fixtures", "credential.json"),
	}

	client, err := config.Client()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assertString(t, testValueConfigPublicKey, client.credential.PublicKey)
	assertString(t, testValueConfigPrivateKey, client.credential.PrivateKey)
}

func assertString(t *testing.T, expected string, value string) {
	if expected != value {
		t.Errorf("expected %q, but got %q", expected, value)
	}
}
