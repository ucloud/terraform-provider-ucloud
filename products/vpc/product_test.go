package vpc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

var vpcTerraformNamespaces = []string{
	"vpc", "subnet", "vpcs", "subnets", "vip", "nat_gateway", "nat_gateways", "sec_group", "sec_groups", "route_table", "network_interface",
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
		"ucloud_route_table",
		"ucloud_route_table_rule",
		"ucloud_route_table_association",
		"ucloud_network_interface",
		"ucloud_sec_group",
		"ucloud_sec_group_rule",
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
		"ucloud_route_table": {
			"name", "tag", "remark", "vpc_id", "vpc_name", "route_table_type", "subnet_count", "subnet_ids", "create_time",
		},
		"ucloud_route_table_rule": {
			"route_table_id", "dst_addr", "nexthop_type", "nexthop_id", "remark",
		},
		"ucloud_route_table_association": {
			"subnet_id", "route_table_id",
		},
		"ucloud_network_interface": {
			"name", "tag", "remark", "vpc_id", "subnet_id", "private_ip", "instance_id", "mac_address", "status", "create_time",
		},
		"ucloud_sec_group": {
			"name", "vpc_id", "tag", "type", "remark", "create_time",
		},
		"ucloud_sec_group_rule": {
			"sec_group_id", "direction", "protocol_type", "dst_port", "ip_range", "rule_action", "priority", "remark",
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

	if field := resources["ucloud_sec_group"].Schema["tag"]; field == nil || !field.Computed {
		t.Error("ucloud_sec_group tag must stay computed")
	}
	if field := resources["ucloud_sec_group"].Schema["vpc_id"]; field == nil || !field.ForceNew {
		t.Error("ucloud_sec_group vpc_id must stay ForceNew")
	}
	// The group is the rule's parent, not one of its attributes: a rule cannot
	// move between groups, so this is the one field that rebuilds the rule.
	if field := resources["ucloud_sec_group_rule"].Schema["sec_group_id"]; field == nil || !field.ForceNew {
		t.Error("ucloud_sec_group_rule sec_group_id must stay ForceNew")
	}
	// A group ID is the one value that decides which group the rule lands in,
	// so a typo here is worth catching before it reaches the API.
	if field := resources["ucloud_sec_group_rule"].Schema["sec_group_id"]; field == nil || field.ValidateFunc == nil {
		t.Error("ucloud_sec_group_rule sec_group_id must validate the group ID")
	}
	// Everything else is edited in place by UpdateSecGroupRule, which addresses
	// the rule by its ID alone. Locking any of these would take away the reason
	// the rule is a resource of its own.
	for _, field := range []string{"direction", "protocol_type", "dst_port", "ip_range", "rule_action", "priority", "remark"} {
		if resources["ucloud_sec_group_rule"].Schema[field].ForceNew {
			t.Errorf("ucloud_sec_group_rule %q must stay updatable", field)
		}
	}

	for _, name := range []string{"ucloud_vpc", "ucloud_subnet", "ucloud_sec_group", "ucloud_sec_group_rule"} {
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
	if resources["ucloud_sec_group_rule"].CustomizeDiff == nil {
		t.Error("ucloud_sec_group_rule CustomizeDiff is not configured")
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
	for value, wantErr := range map[string]bool{
		"":                     false,
		"80":                   false,
		"80,443":               false,
		"443,2000-10000":       false,
		"80,":                  true,
		"0,443":                true,
		"10000-2000":           true,
		"443,2000-10000,70000": true,
	} {
		_, errors := validateSecGroupPortList(value, "dst_port")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("sec group port list validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"":                          true,
		"0.0.0.0/0":                 false,
		"10.0.0.0/8,192.168.0.0/16": false,
		"2001:db8::/32":             false,
		"10.0.0.1/8":                true,
		"0.0.0.0/0,":                true,
		// The three forms are mutually exclusive: a CIDR list, a single
		// security group, or a list of prefix lists.
		"secgroup-abc123":            false,
		"pl-abc123":                  false,
		"pl-abc123,pl-def456":        false,
		"secgroup-abc123,pl-def456":  true,
		"pl-abc123,secgroup-abc123":  true,
		"10.0.0.0/8,secgroup-abc123": true,
		"secgroup-abc123,10.0.0.0/8": true,
		"pl-abc123,10.0.0.0/8":       true,
		// Two groups in one rule would name two sources, so a group stands
		// alone.
		"secgroup-abc123,secgroup-def456": true,
		// The prefix alone names nothing.
		"secgroup-": true,
		"pl-":       true,
	} {
		_, errors := validateSecGroupIPRange(value, "ip_range")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("sec group ip range validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"secgroup-abc123": false,
		"secgroup-":       true,
		"pl-abc123":       true,
		"abc123":          true,
		"":                true,
	} {
		_, errors := validateSecGroupResourceID(value, "sec_group_id")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("sec group id validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	// A port the API fills in on its own for a port independent protocol must
	// not reach the state, or it would diff against the configuration forever.
	for protocol, want := range map[string]string{
		"TCP":    "22",
		"UDP":    "22",
		"ICMP":   "",
		"ICMPv6": "",
		"ALL":    "",
	} {
		if got := secGroupRuleDstPort(vpcapi.SecGroupRuleInfo{ProtocolType: protocol, DstPort: "22"}); got != want {
			t.Errorf("sec group rule destination port for %s = %q, want %q", protocol, got, want)
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

// TestSecGroupRuleParameterMapsTheConfiguration covers the attributes the rule
// parameter carries, and the destination port above all.
//
// A protocol that has no use for a port is sent one anyway. The empty string
// cannot travel: the SDK's request encoder drops it before the request leaves,
// and the API reads the resulting missing parameter as an invalid rule rather
// than as "all ports", rejecting the call with 208605. So the wire spelling of an
// empty port is pinned down here, next to the read side that turns it back into
// the empty string - the two spellings are a pair, and either drifting away from
// the other is silent. The plan-time check that rejects a port on such a rule is
// covered separately, in TestDiffValidateSecGroupRulePort.
func TestSecGroupRuleParameterMapsTheConfiguration(t *testing.T) {
	ruleSchema := resourceUCloudSecGroupRule().Schema

	for protocol, wantPort := range map[string]bool{
		"TCP":    true,
		"UDP":    true,
		"ICMP":   false,
		"ICMPv6": false,
		"ALL":    false,
	} {
		raw := map[string]interface{}{
			"sec_group_id":  "secgroup-abc123",
			"direction":     "Ingress",
			"protocol_type": protocol,
			"ip_range":      "10.0.0.0/8",
			"rule_action":   "Accept",
			"priority":      50,
			"remark":        "",
		}
		if wantPort {
			raw["dst_port"] = "22"
		}

		param := secGroupRuleParameter(schema.TestResourceDataRaw(t, ruleSchema, raw))
		form, err := request.EncodeForm(&vpcapi.CreateSecGroupRuleRequest{
			Rule: []vpcapi.CreateSecGroupRuleParamRule{param},
		})
		if err != nil {
			t.Fatalf("encode sec group rule request: %s", err)
		}
		wantOnWire := "22"
		if !wantPort {
			wantOnWire = secGroupRulePortAll
		}
		if got := form["Rule.0.DstPort"]; got != wantOnWire {
			t.Errorf("sec group rule destination port on the wire for %s = %q, want %q", protocol, got, wantOnWire)
		}
		if got := *param.ProtocolType; got != protocol {
			t.Errorf("sec group rule protocol type = %q, want %q", got, protocol)
		}
		if got := *param.Priority; got != 50 {
			t.Errorf("sec group rule priority for %s = %d, want 50", protocol, got)
		}
	}
}

// Create and update take structurally identical rule parameters, and the update
// path builds its own struct by copying the one the create path builds. Nothing
// else stops an SDK field from being added to one of the two and silently
// dropped by the other, which would make every edit revert whatever the new
// field carries.
func TestSecGroupRuleParamStructsStayInStep(t *testing.T) {
	fieldNames := func(value interface{}) []string {
		valueType := reflect.TypeOf(value)
		names := make([]string, 0, valueType.NumField())
		for index := 0; index < valueType.NumField(); index++ {
			names = append(names, valueType.Field(index).Name)
		}
		return names
	}

	createFields := fieldNames(vpcapi.CreateSecGroupRuleParamRule{})
	updateFields := fieldNames(vpcapi.UpdateSecGroupRuleParamRule{})

	extra := make([]string, 0, 1)
	for _, name := range updateFields {
		if !isStringIn(name, createFields) {
			extra = append(extra, name)
		}
	}
	if len(extra) != 1 || extra[0] != "RuleId" {
		t.Fatalf("UpdateSecGroupRuleParamRule must add exactly RuleId to CreateSecGroupRuleParamRule, got %v", extra)
	}
	for _, name := range createFields {
		if !isStringIn(name, updateFields) {
			t.Fatalf("field %q is sent on create but missing from UpdateSecGroupRuleParamRule, so an edit would silently drop it", name)
		}
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

	// baseURL, when set, points the product client at a local server, so a test
	// can drive the SDK against a canned response instead of the real API.
	baseURL string
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	stub.name = name
	stub.calls++
	config := ucloud.NewConfig()
	config.Region = "cn-test"
	config.ProjectId = "project-test"
	if stub.baseURL != "" {
		config.BaseUrl = stub.baseURL
	}
	credential := auth.NewCredential()
	return constructor(&config, &credential, nil), nil
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}

// TestSecGroupServiceEmptyIdentifiersAreNotFound pins the guards that keep a
// missing identifier from being sent to the API at all. Both the read and the
// delete path lean on isNotFoundError to tell "the group is gone" apart from
// "the call failed", so an empty identifier has to arrive as the provider's own
// not-found error: a read that cannot tell the two apart fails instead of
// clearing the ID, and a delete that cannot tell them apart stops being
// idempotent.
//
// The client is pointed at a local server that fails the test if it is ever
// reached. With the guards in place nothing is sent, and a guard that regresses
// fails here rather than calling the real API.
func TestSecGroupServiceEmptyIdentifiersAreNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Errorf("empty identifier reached the API: %s", request.FormValue("Action"))
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := clientFromMeta(&runtimeStub{baseURL: server.URL})
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}

	if _, err := client.describeSecGroupById(""); err == nil || !isNotFoundError(err) {
		t.Fatalf("describe empty sec group id error = %v, want provider not-found error", err)
	}
	if _, err := client.describeSecGroupRuleById("secgroup-abc123456", ""); err == nil || !isNotFoundError(err) {
		t.Fatalf("describe empty sec group rule id error = %v, want provider not-found error", err)
	}
	if _, err := client.describeSecGroupRuleById("", "rule-abc123456"); err == nil || !isNotFoundError(err) {
		t.Fatalf("describe rule of empty sec group id error = %v, want provider not-found error", err)
	}
}

// TestSecGroupServiceTranslatesMissingGroupAndPropagatesOtherErrors covers the
// retcode translation in describeSecGroupById against a canned backend. A group
// that no longer exists is reported as a server error rather than as an empty
// DataSet, so both codes that mean gone have to become the provider's own
// not-found error; every other retcode has to travel up untouched, so that a
// real failure is never mistaken for a missing group and turned into a cleared
// ID or a successful delete.
func TestSecGroupServiceTranslatesMissingGroupAndPropagatesOtherErrors(t *testing.T) {
	const secGroupID = "secgroup-abc123456"

	tests := []struct {
		name         string
		retCode      int
		message      string
		wantNotFound bool
	}{
		{name: "generic not-found code", retCode: 54002, message: "sec group not found", wantNotFound: true},
		{name: "backend not-found code", retCode: 208704, message: "sec group does not exist", wantNotFound: true},
		{name: "other failure", retCode: 999, message: "internal server error"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if err := request.ParseForm(); err != nil {
					t.Errorf("parse request form: %v", err)
				}
				if got := request.Form.Get("Action"); got != "DescribeSecGroup" {
					t.Errorf("request action = %q, want DescribeSecGroup", got)
				}
				if got := request.Form.Get("SecGroupId.0"); got != secGroupID {
					t.Errorf("request SecGroupId.0 = %q, want %q", got, secGroupID)
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(writer, `{"RetCode":%d,"Message":%q}`, test.retCode, test.message)
			}))
			defer server.Close()

			client, err := clientFromMeta(&runtimeStub{baseURL: server.URL})
			if err != nil {
				t.Fatalf("get product client: %v", err)
			}

			group, err := client.describeSecGroupById(secGroupID)
			if test.wantNotFound {
				if err == nil || !isNotFoundError(err) {
					t.Fatalf("describe missing sec group = (%#v, %v), want provider not-found error", group, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("describe failed sec group = (%#v, nil), want the backend failure to propagate", group)
			}
			if isNotFoundError(err) {
				t.Fatalf("describe failed sec group error = %v, want it to stay a backend failure", err)
			}
			if !strings.Contains(err.Error(), test.message) {
				t.Fatalf("describe failed sec group error = %v, want it to carry %q", err, test.message)
			}
		})
	}
}

// TestSecGroupServiceReadsGroupAndRuleResponse covers the response side. An
// empty DataSet means the group is gone, so it has to read as not found rather
// than as a group with no fields; a returned group is handed back whole, rules
// included. The rule cases are what the composite import ID exists for: a rule
// is found by scanning its parent group, and one absent from that group has to
// read as not found, or a delete would stop being idempotent.
func TestSecGroupServiceReadsGroupAndRuleResponse(t *testing.T) {
	const (
		secGroupID   = "secgroup-abc123456"
		ruleID       = "rule-abc123456"
		absentRuleID = "rule-def789012"
	)

	groupWithRule := `{"RetCode":0,"DataSet":[{"SecGroupId":"` + secGroupID +
		`","Name":"tf-test","Rule":[{"RuleId":"` + ruleID + `","DstPort":"22"}]}]}`

	t.Run("empty data set is not found", func(t *testing.T) {
		client := secGroupClientForBody(t, `{"RetCode":0,"DataSet":[]}`)
		if group, err := client.describeSecGroupById(secGroupID); err == nil || !isNotFoundError(err) {
			t.Fatalf("describe absent sec group = (%#v, %v), want provider not-found error", group, err)
		}
	})

	t.Run("group is returned with its rules", func(t *testing.T) {
		group, err := secGroupClientForBody(t, groupWithRule).describeSecGroupById(secGroupID)
		if err != nil {
			t.Fatalf("describe sec group: %v", err)
		}
		if group.SecGroupId != secGroupID || group.Name != "tf-test" {
			t.Fatalf("described sec group = %#v, want id %q and name tf-test", group, secGroupID)
		}
		if len(group.Rule) != 1 || group.Rule[0].RuleId != ruleID || group.Rule[0].DstPort != "22" {
			t.Fatalf("described sec group rules = %#v, want rule %q on port 22", group.Rule, ruleID)
		}
	})

	t.Run("rule is located by scanning its group", func(t *testing.T) {
		rule, err := secGroupClientForBody(t, groupWithRule).describeSecGroupRuleById(secGroupID, ruleID)
		if err != nil {
			t.Fatalf("describe sec group rule: %v", err)
		}
		if rule.RuleId != ruleID || rule.DstPort != "22" {
			t.Fatalf("described sec group rule = %#v, want rule %q on port 22", rule, ruleID)
		}
	})

	t.Run("rule absent from its group is not found", func(t *testing.T) {
		rule, err := secGroupClientForBody(t, groupWithRule).describeSecGroupRuleById(secGroupID, absentRuleID)
		if err == nil || !isNotFoundError(err) {
			t.Fatalf("describe absent sec group rule = (%#v, %v), want provider not-found error", rule, err)
		}
	})

	t.Run("parent group failure propagates through the rule lookup", func(t *testing.T) {
		rule, err := secGroupClientForBody(t, `{"RetCode":999,"Message":"internal server error"}`).
			describeSecGroupRuleById(secGroupID, ruleID)
		if err == nil {
			t.Fatalf("describe sec group rule = (%#v, nil), want the parent group failure to propagate", rule)
		}
		if isNotFoundError(err) {
			t.Fatalf("describe sec group rule error = %v, want it to stay a backend failure", err)
		}
	})
}

// secGroupClientForBody builds a product client whose VPC API answers every call
// with body, so a test can pin how a response shape is read without a backend.
func secGroupClientForBody(t *testing.T, body string) *productClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, body)
	}))
	t.Cleanup(server.Close)

	client, err := clientFromMeta(&runtimeStub{baseURL: server.URL})
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}
	return client
}
