package ulb

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("lb", "lbs"))); err != nil {
		t.Fatalf("register ulb: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate ulb provider: %v", err)
	}
	for _, name := range []string{
		"ucloud_lb",
		"ucloud_lb_listener",
		"ucloud_lb_attachment",
		"ucloud_lb_rule",
		"ucloud_lb_ssl",
		"ucloud_lb_ssl_attachment",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{
		"ucloud_lbs",
		"ucloud_lb_listeners",
		"ucloud_lb_attachments",
		"ucloud_lb_rules",
		"ucloud_lb_ssls",
	} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceSchemasKeepCompatibilityFields(t *testing.T) {
	resources := New().Registration().Resources
	for name, fields := range map[string][]string{
		"ucloud_lb":                {"internal", "vpc_id", "subnet_id", "charge_type", "name", "tag", "remark", "security_group", "listen_type", "ip_set", "private_ip", "create_time", "expire_time"},
		"ucloud_lb_listener":       {"load_balancer_id", "protocol", "name", "listen_type", "port", "idle_timeout", "method", "persistence_type", "persistence", "health_check_type", "domain", "path", "status"},
		"ucloud_lb_attachment":     {"load_balancer_id", "listener_id", "resource_type", "resource_id", "port", "private_ip", "status"},
		"ucloud_lb_rule":           {"load_balancer_id", "listener_id", "backend_ids", "domain", "path"},
		"ucloud_lb_ssl":            {"name", "private_key", "user_cert", "ca_cert", "create_time"},
		"ucloud_lb_ssl_attachment": {"ssl_id", "load_balancer_id", "listener_id"},
	} {
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

	if got := resources["ucloud_lb"].Schema["tag"].Default; got != defaultTag {
		t.Errorf("lb tag default = %#v, want %q", got, defaultTag)
	}
	if got := resources["ucloud_lb_attachment"].Schema["port"].Default; got != 80 {
		t.Errorf("attachment port default = %#v, want 80", got)
	}
	if got := resources["ucloud_lb_listener"].Schema["load_balancer_id"].ForceNew; !got {
		t.Error("listener load_balancer_id must remain ForceNew")
	}
	if got := resources["ucloud_lb_rule"].Schema["backend_ids"].ForceNew; !got {
		t.Error("rule backend_ids must remain ForceNew")
	}
}

func TestDataSourceSchemasKeepCompatibilityFields(t *testing.T) {
	dataSources := New().Registration().DataSources
	for name, fields := range map[string][]string{
		"ucloud_lbs":            {"ids", "name_regex", "vpc_id", "subnet_id", "output_file", "total_count", "lbs"},
		"ucloud_lb_listeners":   {"ids", "load_balancer_id", "name_regex", "output_file", "total_count", "lb_listeners"},
		"ucloud_lb_attachments": {"ids", "load_balancer_id", "listener_id", "output_file", "total_count", "lb_attachments"},
		"ucloud_lb_rules":       {"ids", "load_balancer_id", "listener_id", "output_file", "total_count", "lb_rules"},
		"ucloud_lb_ssls":        {"ids", "name_regex", "output_file", "total_count", "lb_ssls"},
	} {
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
}

func TestListenerChoicesCompatibility(t *testing.T) {
	if got := availableLBChoices.availableChoices("http"); len(got) != 1 || got[0] != "request_proxy" {
		t.Fatalf("http listen choices = %#v, want [request_proxy]", got)
	}
	if got := availableLBChoices.availableChoices("tcp"); len(got) != 2 || !isStringIn("packets_transmit", got) {
		t.Fatalf("tcp listen choices = %#v, want request_proxy and packets_transmit", got)
	}
	if err := availableLBChoices.validate("udp", "request_proxy"); err != nil {
		t.Fatalf("valid udp listen type rejected: %v", err)
	}
	if err := availableLBChoices.validate("http", "packets_transmit"); err == nil {
		t.Fatal("invalid http listen type was accepted")
	}
}

func TestCompatibilityConvertersAndState(t *testing.T) {
	for input, want := range map[string]string{
		"RequestProxy":       "request_proxy",
		"PacketsTransmit":    "packets_transmit",
		"ConsistentHashPort": "consistent_hash_port",
	} {
		if got := upperCamelConvert(input); got != want {
			t.Errorf("upperCamelConvert(%q) = %q, want %q", input, got, want)
		}
	}
	if got := upperCamelUnconvert("request_proxy"); got != "RequestProxy" {
		t.Errorf("upperCamelUnconvert(request_proxy) = %q", got)
	}
	if got := upperCvt.convert("HTTP"); got != "http" {
		t.Errorf("upperCvt.convert(HTTP) = %q, want http", got)
	}
	if got := stateFuncTag(""); got != defaultTag {
		t.Errorf("empty tag state = %q, want %q", got, defaultTag)
	}

	legacy := &terraform.InstanceState{
		ID: "lb-legacy",
		Attributes: map[string]string{
			"name": "legacy-lb",
			"tag":  "Default",
		},
	}
	state := resourceUCloudLB().Data(legacy).State()
	if state == nil || state.ID != legacy.ID {
		t.Fatal("legacy LB state was not retained")
	}
	if state.Attributes["name"] != legacy.Attributes["name"] || state.Attributes["tag"] != legacy.Attributes["tag"] {
		t.Fatal("legacy LB state attributes changed")
	}
	if _, errors := validateName(strings.Repeat("a", 64), "name"); len(errors) == 0 {
		t.Fatal("overlong LB name was accepted")
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
	return constructor(&config, &auth.Credential{}, nil), nil
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
	if client == nil || client.ulbconn == nil || client.unetconn == nil {
		t.Fatal("product client did not initialize ULB and UNet SDK clients")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}
