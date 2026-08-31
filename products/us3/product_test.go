package us3

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New())); err != nil {
		t.Fatalf("register US3 product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with US3 product: %v", err)
	}
	if provider.ResourcesMap["ucloud_us3_bucket"] == nil {
		t.Fatal("ucloud_us3_bucket is not registered")
	}
	if provider.DataSourcesMap["ucloud_us3_buckets"] == nil {
		t.Fatal("ucloud_us3_buckets is not registered")
	}
}

func TestBucketNameValidationCompatibility(t *testing.T) {
	validate := New().Registration().Resources["ucloud_us3_bucket"].Schema["name"].ValidateFunc
	tests := map[string]struct {
		value   string
		wantErr bool
	}{
		"minimum length":     {value: "abcde1"},
		"below minimum":      {value: "abcd1", wantErr: true},
		"maximum length":     {value: strings.Repeat("a", 64)},
		"above maximum":      {value: strings.Repeat("a", 65), wantErr: true},
		"uppercase":          {value: "TF-acc-us3-bucket-basic", wantErr: true},
		"reserved www":       {value: "www-example-bucket", wantErr: true},
		"reserved cn-bj":     {value: "cn-bj-example-bucket", wantErr: true},
		"reserved hk":        {value: "hk-example-bucket", wantErr: true},
		"starts with hyphen": {value: "-example-bucket", wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, errors := validate(tc.value, "name")
			if got := len(errors) > 0; got != tc.wantErr {
				t.Fatalf("validation errors = %v, wantErr = %t", errors, tc.wantErr)
			}
		})
	}
}

func TestBucketTypeValidationCompatibility(t *testing.T) {
	validate := New().Registration().Resources["ucloud_us3_bucket"].Schema["type"].ValidateFunc
	for value, wantErr := range map[string]bool{
		"public":  false,
		"private": false,
		"PUBLIC":  true,
		"":        true,
	} {
		_, errors := validate(value, "type")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

func TestBucketTagCompatibility(t *testing.T) {
	tag := New().Registration().Resources["ucloud_us3_bucket"].Schema["tag"]
	if tag.Default != defaultTag {
		t.Fatalf("tag default = %#v, want %q", tag.Default, defaultTag)
	}
	for value, want := range map[string]string{"": defaultTag, "team-a": "team-a"} {
		if got := tag.StateFunc(value); got != want {
			t.Errorf("tag StateFunc(%q) = %q, want %q", value, got, want)
		}
	}
	for value, wantErr := range map[string]bool{
		"":                      false,
		"team-a":                false,
		strings.Repeat("a", 63): false,
		strings.Repeat("a", 64): true,
		"team/a":                true,
	} {
		_, errors := tag.ValidateFunc(value, "tag")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("tag validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

func TestBucketResourceAcceptsLegacyState(t *testing.T) {
	fixture, err := os.ReadFile("test-fixtures/state-v0.json")
	if err != nil {
		t.Fatalf("read legacy state fixture: %v", err)
	}
	var legacy terraform.InstanceState
	if err := json.Unmarshal(fixture, &legacy); err != nil {
		t.Fatalf("decode legacy state fixture: %v", err)
	}

	state := New().Registration().Resources["ucloud_us3_bucket"].Data(&legacy).State()
	if state == nil {
		t.Fatal("legacy state was dropped")
	}
	if state.ID != legacy.ID {
		t.Fatalf("state ID = %q, want %q", state.ID, legacy.ID)
	}
	for _, name := range []string{"name", "type", "tag", "create_time", "source_domain_names.#", "source_domain_names.0"} {
		if state.Attributes[name] != legacy.Attributes[name] {
			t.Errorf("state attribute %q = %q, want %q", name, state.Attributes[name], legacy.Attributes[name])
		}
	}
}
