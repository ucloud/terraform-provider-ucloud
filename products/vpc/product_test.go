package vpc

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

var vpcTerraformNamespaces = []string{
	"vpc", "subnet", "vpcs", "subnets", "vip", "nat_gateway", "nat_gateways", "sec_groups",
}

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces(vpcTerraformNamespaces...))); err != nil {
		t.Fatalf("register vpc: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate vpc provider: %v", err)
	}
	for _, name := range []string{
		"ucloud_vpc",
		"ucloud_subnet",
		"ucloud_vpc_peering_connection",
		"ucloud_vip",
		"ucloud_nat_gateway",
		"ucloud_nat_gateway_rule",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_vpcs", "ucloud_subnets", "ucloud_nat_gateways", "ucloud_sec_groups"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceSchemasKeepCompatibility(t *testing.T) {
	resources := New().Registration().Resources
	wantFields := map[string][]string{
		"ucloud_vpc": {
			"name", "cidr_blocks", "tag", "remark", "network_info", "update_time", "create_time",
		},
		"ucloud_subnet": {
			"cidr_block", "vpc_id", "name", "tag", "remark", "create_time",
		},
		"ucloud_vpc_peering_connection": {
			"vpc_id", "peer_vpc_id", "peer_project_id", "peer_region",
		},
		"ucloud_vip": {
			"vpc_id", "subnet_id", "name", "tag", "remark", "ip_address", "create_time",
		},
		"ucloud_nat_gateway": {
			"vpc_id", "eip_id", "subnet_ids", "security_group", "enable_white_list", "white_list", "name", "tag", "remark", "create_time",
		},
		"ucloud_nat_gateway_rule": {
			"nat_gateway_id", "protocol", "src_eip_id", "src_port_range", "dst_ip", "dst_port_range", "name",
		},
	}
	for name, fields := range wantFields {
		resource := resources[name]
		if resource == nil {
			t.Fatalf("resource %q is missing", name)
		}
		for _, field := range fields {
			if resource.Schema[field] == nil {
				t.Errorf("resource %q is missing field %q", name, field)
			}
		}
	}

	for _, name := range []string{"ucloud_vpc", "ucloud_subnet"} {
		if resources[name].Importer == nil || resources[name].Importer.State == nil {
			t.Errorf("resource %q importer is not configured", name)
		}
	}
	if resources["ucloud_vpc"].CustomizeDiff == nil {
		t.Error("ucloud_vpc CustomizeDiff is not configured")
	}
	if resources["ucloud_nat_gateway_rule"].CustomizeDiff == nil {
		t.Error("ucloud_nat_gateway_rule CustomizeDiff is not configured")
	}
	if resources["ucloud_vpc"].Schema["tag"].Default != defaultTag {
		t.Errorf("vpc tag default = %#v, want %q", resources["ucloud_vpc"].Schema["tag"].Default, defaultTag)
	}
	if resources["ucloud_nat_gateway"].Schema["name"].ValidateFunc == nil {
		t.Error("nat gateway name validator is missing")
	}
}

func TestDataSourceSchemasKeepCompatibility(t *testing.T) {
	dataSources := New().Registration().DataSources
	wantFields := map[string][]string{
		"ucloud_vpcs":         {"ids", "name_regex", "tag", "output_file", "total_count", "vpcs"},
		"ucloud_subnets":      {"ids", "name_regex", "vpc_id", "tag", "output_file", "total_count", "subnets"},
		"ucloud_nat_gateways": {"name_regex", "ids", "output_file", "total_count", "nat_gateways"},
		"ucloud_sec_groups":   {"ids", "name_regex", "vpc_id", "output_file", "total_count", "sec_groups"},
	}
	for name, fields := range wantFields {
		dataSource := dataSources[name]
		if dataSource == nil {
			t.Fatalf("data source %q is missing", name)
		}
		for _, field := range fields {
			if dataSource.Schema[field] == nil {
				t.Errorf("data source %q is missing field %q", name, field)
			}
		}
	}
	for _, name := range []string{"vpcs", "subnets", "nat_gateways", "sec_groups"} {
		field := dataSources["ucloud_"+name].Schema[name]
		if field.Type != schema.TypeList || !field.Computed {
			t.Errorf("data source %q output must remain computed TypeList", name)
		}
	}
}

func TestCompatibilityValidationAndPureLogic(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"10.0.0.0/8":       false,
		"172.16.0.0/12":    false,
		"192.168.0.0/16":   false,
		"192.168.1.1/24":   true,
		"2001:db8::/32":    true,
		"10.0.0.0/30":      true,
		"10.0.0.0/8/extra": true,
	} {
		_, errors := validateCIDRBlock(value, "cidr_block")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("CIDR validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"80":     false,
		"80-443": false,
		"0":      true,
		"443-80": true,
		"80-":    true,
	} {
		_, errors := validatePortRange(value, "port_range")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("port validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	if got := stateFuncTag(""); got != defaultTag {
		t.Errorf("empty tag state = %q, want %q", got, defaultTag)
	}
	if got := upperCvt.convert("TCP"); got != "tcp" {
		t.Errorf("upperCvt.convert(TCP) = %q, want tcp", got)
	}
	if got := upperCvt.unconvert("udp"); got != "UDP" {
		t.Errorf("upperCvt.unconvert(udp) = %q, want UDP", got)
	}
	for input, want := range map[string][2]string{
		"cn-bj@project":       {"cn-bj", "project"},
		"cn-sh@other@ignored": {"cn-sh", "other"},
	} {
		region, project, err := parseVPCPeerDstType(input)
		if err != nil || region != want[0] || project != want[1] {
			t.Errorf("parseVPCPeerDstType(%q) = (%q, %q, %v)", input, region, project, err)
		}
	}
	if _, _, err := parseVPCPeerDstType("invalid"); err == nil {
		t.Fatal("invalid peer destination type was accepted")
	}
	if err := diffSupressVPCNetworkUpdate(schema.NewSet(hashCIDR, []interface{}{"10.0.0.0/8"}), schema.NewSet(hashCIDR, []interface{}{"10.0.0.0/8", "172.16.0.0/12"}), nil); err != nil {
		t.Fatalf("create-only network change rejected: %v", err)
	}
	if err := diffSupressVPCNetworkUpdate(schema.NewSet(hashCIDR, []interface{}{"10.0.0.0/8", "172.16.0.0/12"}), schema.NewSet(hashCIDR, []interface{}{"10.0.0.0/8", "192.168.0.0/16"}), nil); err == nil {
		t.Fatal("simultaneous network add and delete was accepted")
	}
}

func TestLegacyStateAndRuntimeCompatibility(t *testing.T) {
	legacy := &terraform.InstanceState{
		ID: "vpc-legacy",
		Attributes: map[string]string{
			"name": "legacy-vpc",
			"tag":  defaultTag,
		},
	}
	state := resourceUCloudVPC().Data(legacy).State()
	if state == nil || state.ID != legacy.ID {
		t.Fatal("legacy VPC state was not retained")
	}
	if state.Attributes["name"] != legacy.Attributes["name"] || state.Attributes["tag"] != legacy.Attributes["tag"] {
		t.Fatal("legacy VPC state attributes changed")
	}
	if _, errors := validateName(strings.Repeat("a", 64), "name"); len(errors) == 0 {
		t.Fatal("overlong VPC name was accepted")
	}

	runtime := &runtimeStub{}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}
	if runtime.name != Name || runtime.calls != 1 {
		t.Fatalf("runtime call = (%q, %d), want (%q, 1)", runtime.name, runtime.calls, Name)
	}
	if client.region != "cn-test" || client.projectId != "project-test" {
		t.Fatalf("client identity = (%q, %q), want (cn-test, project-test)", client.region, client.projectId)
	}
	if client.vpcconn == nil || client.unetconn == nil {
		t.Fatal("product client did not initialize VPC and UNet SDK clients")
	}
}

type runtimeStub struct {
	name  string
	calls int
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	stub.name = name
	stub.calls++
	config := ucloud.NewConfig()
	config.Region = "cn-test"
	config.ProjectId = "project-test"
	credential := auth.NewCredential()
	return constructor(&config, &credential, nil), nil
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}
