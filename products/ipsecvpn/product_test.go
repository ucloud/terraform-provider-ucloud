package ipsecvpn

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type runtimeStub struct {
	name  string
	calls int
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	stub.name = name
	stub.calls++
	config := ucloud.NewConfig()
	return constructor(&config, &auth.Credential{}, nil), nil
}

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("vpn"))); err != nil {
		t.Fatalf("register ipsecvpn: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate ipsecvpn provider: %v", err)
	}

	for _, name := range []string{
		"ucloud_vpn_gateway",
		"ucloud_vpn_customer_gateway",
		"ucloud_vpn_connection",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{
		"ucloud_vpn_gateways",
		"ucloud_vpn_customer_gateways",
		"ucloud_vpn_connections",
	} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceSchemasKeepCompatibility(t *testing.T) {
	resources := New().Registration().Resources
	for name, fields := range map[string][]string{
		"ucloud_vpn_gateway":          {"vpc_id", "grade", "eip_id", "charge_type", "duration", "name", "tag", "remark", "create_time", "expire_time"},
		"ucloud_vpn_customer_gateway": {"ip_address", "name", "tag", "remark", "create_time"},
		"ucloud_vpn_connection":       {"vpn_gateway_id", "customer_gateway_id", "vpc_id", "name", "tag", "remark", "ike_config", "ipsec_config", "create_time"},
	} {
		resource := resources[name]
		if resource == nil {
			t.Fatalf("resource %q is missing", name)
		}
		if resource.Create == nil || resource.Read == nil || resource.Delete == nil {
			t.Errorf("resource %q is missing CRUD handlers", name)
		}
		if resource.Importer == nil || resource.Importer.State == nil {
			t.Errorf("resource %q importer is not configured", name)
		}
		for _, field := range fields {
			if resource.Schema[field] == nil {
				t.Errorf("resource %q is missing field %q", name, field)
			}
		}
	}

	gateway := resources["ucloud_vpn_gateway"]
	if gateway.Update == nil {
		t.Fatal("vpn gateway update handler is missing")
	}
	if gateway.Schema["vpc_id"].ForceNew != true || gateway.Schema["eip_id"].ForceNew != true {
		t.Fatal("vpn gateway immutable identifiers must remain ForceNew")
	}
	if gateway.Schema["tag"].Default != defaultTag {
		t.Fatalf("vpn gateway tag default = %#v, want %q", gateway.Schema["tag"].Default, defaultTag)
	}
	if got := gateway.Schema["tag"].StateFunc(""); got != defaultTag {
		t.Fatalf("vpn gateway empty tag state = %q, want %q", got, defaultTag)
	}
	if gateway.Schema["remark"].Optional != true || gateway.Schema["remark"].Computed != true || gateway.Schema["remark"].ForceNew != true {
		t.Fatal("vpn gateway remark flags changed")
	}

	customerGateway := resources["ucloud_vpn_customer_gateway"]
	if customerGateway.Schema["ip_address"].Required != true || customerGateway.Schema["ip_address"].ForceNew != true {
		t.Fatal("customer gateway ip_address flags changed")
	}
	if customerGateway.Schema["remark"].Optional != true || customerGateway.Schema["remark"].Computed != true || customerGateway.Schema["remark"].ForceNew != true {
		t.Fatal("customer gateway remark flags changed")
	}

	connection := resources["ucloud_vpn_connection"]
	for _, field := range []string{"vpn_gateway_id", "customer_gateway_id", "vpc_id"} {
		if !connection.Schema[field].Required || !connection.Schema[field].ForceNew {
			t.Errorf("connection field %q must remain required and ForceNew", field)
		}
	}
	if connection.Schema["ike_config"].Type != schema.TypeList || !connection.Schema["ike_config"].Required || connection.Schema["ike_config"].MaxItems != 1 {
		t.Fatal("ike_config schema changed")
	}
	if connection.Schema["ipsec_config"].Type != schema.TypeList || !connection.Schema["ipsec_config"].Required || connection.Schema["ipsec_config"].MaxItems != 1 {
		t.Fatal("ipsec_config schema changed")
	}
	ike, ok := connection.Schema["ike_config"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ike_config element type = %T, want *schema.Resource", connection.Schema["ike_config"].Elem)
	}
	for field, want := range map[string]interface{}{
		"ike_version":              "ikev1",
		"exchange_mode":            "main",
		"encryption_algorithm":     "aes128",
		"authentication_algorithm": "sha1",
		"dh_group":                 "15",
		"sa_life_time":             86400,
	} {
		if got := ike.Schema[field].Default; got != want {
			t.Errorf("ike_config.%s default = %#v, want %#v", field, got, want)
		}
	}
	ipsec, ok := connection.Schema["ipsec_config"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ipsec_config element type = %T, want *schema.Resource", connection.Schema["ipsec_config"].Elem)
	}
	for field, want := range map[string]interface{}{
		"protocol":                 "esp",
		"encryption_algorithm":     "aes128",
		"authentication_algorithm": "sha1",
		"sa_life_time":             3600,
		"pfs_dh_group":             "disable",
	} {
		if got := ipsec.Schema[field].Default; got != want {
			t.Errorf("ipsec_config.%s default = %#v, want %#v", field, got, want)
		}
	}
}

func TestDataSourceSchemasKeepCompatibility(t *testing.T) {
	dataSources := New().Registration().DataSources
	for name, fields := range map[string][]string{
		"ucloud_vpn_gateways":          {"name_regex", "ids", "tag", "vpc_id", "output_file", "total_count", "vpn_gateways"},
		"ucloud_vpn_customer_gateways": {"name_regex", "ids", "tag", "output_file", "total_count", "vpn_customer_gateways"},
		"ucloud_vpn_connections":       {"name_regex", "ids", "tag", "output_file", "total_count", "vpn_connections"},
	} {
		dataSource := dataSources[name]
		if dataSource == nil {
			t.Fatalf("data source %q is missing", name)
		}
		if dataSource.Read == nil {
			t.Errorf("data source %q read handler is missing", name)
		}
		for _, field := range fields {
			if dataSource.Schema[field] == nil {
				t.Errorf("data source %q is missing field %q", name, field)
			}
		}
	}

	gateways := dataSources["ucloud_vpn_gateways"].Schema["vpn_gateways"]
	if gateways.Type != schema.TypeList || !gateways.Computed {
		t.Fatal("vpn_gateways must remain a computed TypeList")
	}
	gateway, ok := gateways.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("vpn_gateways element type = %T, want *schema.Resource", gateways.Elem)
	}
	for _, field := range []string{"id", "grade", "name", "remark", "tag", "charge_type", "auto_renew", "vpc_id", "create_time", "expire_time", "ip_set"} {
		if gateway.Schema[field] == nil || !gateway.Schema[field].Computed {
			t.Errorf("vpn_gateways nested field %q must be computed", field)
		}
	}

	customerGateways := dataSources["ucloud_vpn_customer_gateways"].Schema["vpn_customer_gateways"]
	if customerGateways.Type != schema.TypeList || !customerGateways.Computed {
		t.Fatal("vpn_customer_gateways must remain a computed TypeList")
	}

	connections := dataSources["ucloud_vpn_connections"].Schema["vpn_connections"]
	if connections.Type != schema.TypeList || !connections.Computed {
		t.Fatal("vpn_connections must remain a computed TypeList")
	}
	connection, ok := connections.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("vpn_connections element type = %T, want *schema.Resource", connections.Elem)
	}
	for _, field := range []string{"id", "name", "remark", "tag", "vpn_gateway_id", "customer_gateway_id", "vpc_id", "create_time", "ike_config", "ipsec_config"} {
		if connection.Schema[field] == nil || !connection.Schema[field].Computed {
			t.Errorf("vpn_connections nested field %q must be computed", field)
		}
	}
	for _, field := range []string{"local_subnet_ids", "remote_subnets", "protocol", "encryption_algorithm", "authentication_algorithm", "sa_life_time", "sa_life_time_bytes", "pfs_dh_group"} {
		if connection.Schema["ipsec_config"].Elem.(*schema.Resource).Schema[field] == nil {
			t.Errorf("vpn_connections ipsec_config field %q is missing", field)
		}
	}
}

func TestDataSourceStateCompatibility(t *testing.T) {
	tests := []struct {
		name string
		data *schema.Resource
		save func(*schema.ResourceData) error
		list string
	}{
		{
			name: "gateways",
			data: dateSourceUCloudVPNGateways(),
			save: func(d *schema.ResourceData) error {
				return dataSourceUCloudVPNGatewaysSave(d, nil)
			},
			list: "vpn_gateways",
		},
		{
			name: "customer gateways",
			data: dateSourceUCloudVPNCustomerGateways(),
			save: func(d *schema.ResourceData) error {
				return dataSourceUCloudVPNCustomerGatewaysSave(d, nil)
			},
			list: "vpn_customer_gateways",
		},
		{
			name: "connections",
			data: dateSourceUCloudVPNConnections(),
			save: func(d *schema.ResourceData) error {
				return dataSourceUCloudVPNConnectionsSave(d, nil)
			},
			list: "vpn_connections",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, test.data.Schema, map[string]interface{}{})
			if err := test.save(d); err != nil {
				t.Fatalf("save empty %s: %v", test.name, err)
			}
			if got, want := d.Id(), hashStringArray(nil); got != want {
				t.Errorf("empty %s ID = %q, want %q", test.name, got, want)
			}
			if got := d.Get("total_count").(int); got != 0 {
				t.Errorf("empty %s total_count = %d, want 0", test.name, got)
			}
			if got := d.Get(test.list).([]interface{}); len(got) != 0 {
				t.Errorf("empty %s result length = %d, want 0", test.name, len(got))
			}
		})
	}
}

func TestCompatibilityConvertersAndValidators(t *testing.T) {
	for input, want := range map[string]string{
		"Standard": "standard",
		"Enhanced": "enhanced",
		"Month":    "month",
		"Dynamic":  "dynamic",
	} {
		if got := upperCamelCvt.convert(input); got != want {
			t.Errorf("upperCamelCvt.convert(%q) = %q, want %q", input, got, want)
		}
	}
	for input, want := range map[string]string{
		"standard": "Standard",
		"enhanced": "Enhanced",
		"month":    "Month",
		"dynamic":  "Dynamic",
	} {
		if got := upperCamelCvt.unconvert(input); got != want {
			t.Errorf("upperCamelCvt.unconvert(%q) = %q, want %q", input, got, want)
		}
	}
	if got := vpnAutoCvt.convert("Auto"); got != "auto" {
		t.Errorf("vpnAutoCvt.convert(Auto) = %q, want auto", got)
	}
	if got := vpnDisableCvt.unconvert("disable"); got != "Disable" {
		t.Errorf("vpnDisableCvt.unconvert(disable) = %q, want Disable", got)
	}
	if got := vpnIkeVersionCvt.convert("IKE V1"); got != "ikev1" {
		t.Errorf("vpnIkeVersionCvt.convert(IKE V1) = %q, want ikev1", got)
	}
	if got := boolCamelCvt.unconvert("Yes"); !got {
		t.Error("boolCamelCvt.unconvert(Yes) = false, want true")
	}

	for value, wantErr := range map[string]bool{
		"validKey_123!": false,
		"":              true,
		"bad key":       true,
	} {
		_, errors := validateVPNPreSharedKey(value, "pre_shared_key")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("pre_shared_key validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"auto":  false,
		"Auto":  true,
		"":      true,
		"local": false,
	} {
		_, errors := validateVpnAuto(value, "local_id")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("local_id validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"192.168.0.0/16": false,
		"172.16.0.0/12":  false,
		"10.0.0.0/8":     false,
		"192.168.1.1/24": true,
		"8.8.8.0/24":     true,
	} {
		_, errors := validateCIDRBlock(value, "remote_subnets")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("remote_subnets validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

func TestClientFromMetaUsesProductRuntime(t *testing.T) {
	stub := &runtimeStub{}
	client, err := clientFromMeta(stub)
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}
	if stub.name != Name || stub.calls != 1 {
		t.Fatalf("runtime call = (%q, %d), want (%q, 1)", stub.name, stub.calls, Name)
	}
	if client == nil {
		t.Fatal("product runtime returned a nil IPSecVPN client")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}
