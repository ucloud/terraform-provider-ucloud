package udpn

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New())); err != nil {
		t.Fatalf("register UDPN product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UDPN product: %v", err)
	}
	if provider.ResourcesMap["ucloud_udpn_connection"] == nil {
		t.Fatal("ucloud_udpn_connection is not registered")
	}
}

func TestConnectionSchemaCompatibility(t *testing.T) {
	connection := New().Registration().Resources["ucloud_udpn_connection"]
	if connection == nil {
		t.Fatal("UDPN connection resource is nil")
	}
	if connection.Create == nil || connection.Read == nil || connection.Update == nil || connection.Delete == nil {
		t.Fatal("UDPN connection CRUD callbacks are incomplete")
	}
	if connection.Importer == nil || connection.Importer.State == nil {
		t.Fatal("UDPN connection importer is missing")
	}
	if connection.CustomizeDiff == nil {
		t.Fatal("UDPN connection CustomizeDiff is missing")
	}

	checks := map[string]struct {
		typeValue schema.ValueType
		optional  bool
		required  bool
		computed  bool
		forceNew  bool
	}{
		"bandwidth":   {typeValue: schema.TypeInt, optional: true},
		"charge_type": {typeValue: schema.TypeString, optional: true, forceNew: true},
		"duration":    {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"peer_region": {typeValue: schema.TypeString, required: true, forceNew: true},
		"create_time": {typeValue: schema.TypeString, computed: true},
		"expire_time": {typeValue: schema.TypeString, computed: true},
	}
	if len(connection.Schema) != len(checks) {
		t.Fatalf("schema has %d fields, want %d", len(connection.Schema), len(checks))
	}
	for name, want := range checks {
		field := connection.Schema[name]
		if field == nil {
			t.Fatalf("schema field %q is missing", name)
		}
		if field.Type != want.typeValue || field.Optional != want.optional || field.Required != want.required ||
			field.Computed != want.computed || field.ForceNew != want.forceNew {
			t.Errorf("schema field %q = type=%v optional=%t required=%t computed=%t forceNew=%t",
				name, field.Type, field.Optional, field.Required, field.Computed, field.ForceNew)
		}
	}

	if got := connection.Schema["bandwidth"].Default; got != 2 {
		t.Errorf("bandwidth default = %#v, want 2", got)
	}
	if got := connection.Schema["charge_type"].Default; got != "month" {
		t.Errorf("charge_type default = %#v, want month", got)
	}
}

func TestConnectionSchemaValidationCompatibility(t *testing.T) {
	connection := New().Registration().Resources["ucloud_udpn_connection"]

	for value, wantErr := range map[int]bool{1: true, 2: false, 1000: false, 1001: true} {
		_, errors := connection.Schema["bandwidth"].ValidateFunc(value, "bandwidth")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("bandwidth validation for %d = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[int]bool{-1: true, 0: false, 9: false, 10: true} {
		_, errors := connection.Schema["duration"].ValidateFunc(value, "duration")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("duration validation for %d = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"year":    false,
		"month":   false,
		"dynamic": false,
		"YEAR":    true,
		"":        true,
	} {
		_, errors := connection.Schema["charge_type"].ValidateFunc(value, "charge_type")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("charge_type validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

func TestClientUsesUDPNRuntimeAndLongTimeout(t *testing.T) {
	config := &ucloud.Config{Region: "cn-bj2", Timeout: time.Second}
	client, ok := newClient(config, &auth.Credential{}, nil).(*udpn.UDPNClient)
	if !ok {
		t.Fatalf("newClient returned %T, want *udpn.UDPNClient", client)
	}
	if got := client.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UDPN client timeout = %s, want 1m0s", got)
	}
	if got := providerRegion(client); got != "cn-bj2" {
		t.Fatalf("provider region = %q, want cn-bj2", got)
	}
}

func TestPeerRegionValidationUsesProductRuntime(t *testing.T) {
	runtime := &testRuntime{config: &ucloud.Config{Region: "cn-bj2"}}
	if err := diffValidateUDPNPeerRegion("", "cn-sh2", runtime); err != nil {
		t.Fatalf("validate peer region: %v", err)
	}
	if err := diffValidateUDPNPeerRegion("", "cn-bj2", runtime); err == nil {
		t.Fatal("expected provider region to be rejected as peer region")
	} else if !strings.Contains(err.Error(), "cn-bj2") {
		t.Fatalf("peer region error = %v, want provider region", err)
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("clientFromMeta accepted a non-runtime value")
	}
}

func TestDescribeDPNByIDEmptyIsNotFound(t *testing.T) {
	client := newClient(&ucloud.Config{Region: "cn-bj2"}, &auth.Credential{}, nil).(*udpn.UDPNClient)
	if _, err := describeDPNById(client, ""); err == nil || !isNotFoundError(err) {
		t.Fatalf("describe empty UDPN id error = %v, want provider not-found error", err)
	}
}

func TestChargeTypeConversionCompatibility(t *testing.T) {
	for api, terraform := range map[string]string{
		"Year":          "year",
		"Month":         "month",
		"Dynamic":       "dynamic",
		"CreateUDBFail": "create_udb_fail",
	} {
		if got := upperCamelConvert(api); got != terraform {
			t.Errorf("upperCamelConvert(%q) = %q, want %q", api, got, terraform)
		}
	}
	for terraform, api := range map[string]string{
		"year":            "Year",
		"month":           "Month",
		"dynamic":         "Dynamic",
		"create_udb_fail": "CreateUdbFail",
	} {
		if got := upperCamelUnconvert(terraform); got != api {
			t.Errorf("upperCamelUnconvert(%q) = %q, want %q", terraform, got, api)
		}
	}
}

type testRuntime struct {
	config *ucloud.Config
}

var _ product.RuntimeV1 = (*testRuntime)(nil)

func (runtime *testRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	if name != Name {
		return nil, fmt.Errorf("unexpected product %q", name)
	}
	return constructor(runtime.config, &auth.Credential{}, nil), nil
}
