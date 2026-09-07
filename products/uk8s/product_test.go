package uk8s

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New())); err != nil {
		t.Fatalf("register UK8S product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UK8S product: %v", err)
	}

	for _, name := range []string{"ucloud_uk8s_cluster", "ucloud_uk8s_node"} {
		if provider.ResourcesMap[name] == nil {
			t.Fatalf("%s is not registered", name)
		}
	}
	if len(provider.DataSourcesMap) != 0 {
		t.Fatalf("unexpected UK8S data sources: %#v", provider.DataSourcesMap)
	}
}

func TestClusterSchemaCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_uk8s_cluster"]
	assertResourceCallbacksAndTimeouts(t, resource, "ucloud_uk8s_cluster")

	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"service_cidr":               {typeValue: schema.TypeString, required: true, forceNew: true},
		"vpc_id":                     {typeValue: schema.TypeString, required: true, forceNew: true},
		"subnet_id":                  {typeValue: schema.TypeString, required: true, forceNew: true},
		"password":                   {typeValue: schema.TypeString, required: true, forceNew: true, sensitive: true},
		"name":                       {typeValue: schema.TypeString, optional: true, computed: true},
		"user_data":                  {typeValue: schema.TypeString, optional: true, forceNew: true},
		"init_script":                {typeValue: schema.TypeString, optional: true, forceNew: true},
		"charge_type":                {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"duration":                   {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"k8s_version":                {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"enable_external_api_server": {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"delete_disks_with_instance": {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"kube_proxy":                 {typeValue: schema.TypeList, optional: true},
		"master":                     {typeValue: schema.TypeList, required: true},
		"status":                     {typeValue: schema.TypeString, computed: true},
		"create_time":                {typeValue: schema.TypeString, computed: true},
		"api_server":                 {typeValue: schema.TypeString, computed: true},
		"external_api_server":        {typeValue: schema.TypeString, computed: true},
		"pod_cidr":                   {typeValue: schema.TypeString, computed: true},
		"image_id":                   {typeValue: schema.TypeString, optional: true, forceNew: true},
	}
	assertSchemaFields(t, resource, wantFields)

	if got := resource.Schema["kube_proxy"].MinItems; got != 1 {
		t.Errorf("kube_proxy MinItems = %d, want 1", got)
	}
	if got := resource.Schema["kube_proxy"].MaxItems; got != 1 {
		t.Errorf("kube_proxy MaxItems = %d, want 1", got)
	}
	master, ok := resource.Schema["master"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("master element type = %T, want *schema.Resource", resource.Schema["master"].Elem)
	}
	if got := master.Schema["availability_zones"].MinItems; got != 3 {
		t.Errorf("master availability_zones MinItems = %d, want 3", got)
	}
	if got := master.Schema["availability_zones"].MaxItems; got != 3 {
		t.Errorf("master availability_zones MaxItems = %d, want 3", got)
	}
	if master.Schema["min_cpu_platform"].Default != "Intel/Auto" {
		t.Errorf("master min_cpu_platform default = %#v, want Intel/Auto", master.Schema["min_cpu_platform"].Default)
	}
}

func TestNodeSchemaCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_uk8s_node"]
	assertResourceCallbacksAndTimeouts(t, resource, "ucloud_uk8s_node")

	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"availability_zone":          {typeValue: schema.TypeString, required: true, forceNew: true},
		"cluster_id":                 {typeValue: schema.TypeString, required: true, forceNew: true},
		"image_id":                   {typeValue: schema.TypeString, optional: true, forceNew: true},
		"password":                   {typeValue: schema.TypeString, required: true, forceNew: true, sensitive: true},
		"instance_type":              {typeValue: schema.TypeString, required: true, forceNew: true},
		"charge_type":                {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"duration":                   {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"boot_disk_type":             {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"data_disk_size":             {typeValue: schema.TypeInt, optional: true, computed: true},
		"data_disk_type":             {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"isolation_group":            {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"subnet_id":                  {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"user_data":                  {typeValue: schema.TypeString, optional: true, forceNew: true},
		"init_script":                {typeValue: schema.TypeString, optional: true, forceNew: true},
		"delete_disks_with_instance": {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"disable_schedule_on_create": {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"min_cpu_platform":           {typeValue: schema.TypeString, optional: true, forceNew: true},
		"status":                     {typeValue: schema.TypeString, computed: true},
		"ip_set":                     {typeValue: schema.TypeList, computed: true},
		"create_time":                {typeValue: schema.TypeString, computed: true},
		"expire_time":                {typeValue: schema.TypeString, computed: true},
	}
	assertSchemaFields(t, resource, wantFields)

	if resource.Schema["min_cpu_platform"].Default != "Intel/Auto" {
		t.Errorf("min_cpu_platform default = %#v, want Intel/Auto", resource.Schema["min_cpu_platform"].Default)
	}
	nested, ok := resource.Schema["ip_set"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ip_set element type = %T, want *schema.Resource", resource.Schema["ip_set"].Elem)
	}
	for _, name := range []string{"ip", "internet_type"} {
		if nested.Schema[name] == nil || !nested.Schema[name].Computed {
			t.Errorf("ip_set nested field %q must be computed", name)
		}
	}
}

func assertResourceCallbacksAndTimeouts(t *testing.T, resource *schema.Resource, name string) {
	t.Helper()
	if resource == nil {
		t.Fatalf("%s resource is nil", name)
	}
	if resource.Create == nil || resource.Read == nil || resource.Update == nil || resource.Delete == nil {
		t.Fatalf("%s CRUD callbacks are incomplete", name)
	}
	if resource.CustomizeDiff == nil {
		t.Fatalf("%s CustomizeDiff is missing", name)
	}
	if resource.Timeouts == nil {
		t.Fatalf("%s timeouts are missing", name)
	}
	for timeoutName, want := range map[string]time.Duration{
		"create": 30 * time.Minute,
		"update": 20 * time.Minute,
		"delete": 10 * time.Minute,
	} {
		var got *time.Duration
		switch timeoutName {
		case "create":
			got = resource.Timeouts.Create
		case "update":
			got = resource.Timeouts.Update
		case "delete":
			got = resource.Timeouts.Delete
		}
		if got == nil || *got != want {
			t.Errorf("%s timeout = %v, want %s", timeoutName, got, want)
		}
	}
}

func assertSchemaFields(t *testing.T, resource *schema.Resource, wantFields map[string]struct {
	typeValue schema.ValueType
	required  bool
	optional  bool
	computed  bool
	forceNew  bool
	sensitive bool
}) {
	t.Helper()
	if len(resource.Schema) != len(wantFields) {
		t.Fatalf("schema has %d fields, want %d", len(resource.Schema), len(wantFields))
	}
	for name, want := range wantFields {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("schema field %q is missing", name)
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Optional != want.optional ||
			field.Computed != want.computed || field.ForceNew != want.forceNew || field.Sensitive != want.sensitive {
			t.Errorf("schema field %q = type=%v required=%t optional=%t computed=%t forceNew=%t sensitive=%t",
				name, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew, field.Sensitive)
		}
	}
}

func TestClientPreservesLegacyHTTPTimeout(t *testing.T) {
	config := &ucloud.Config{Region: "cn-bj2", Timeout: time.Second}
	client, ok := newClient(config, &auth.Credential{}, nil).(*sdkuk8s.UK8SClient)
	if !ok {
		t.Fatalf("newClient() returned %T, want *uk8s.UK8SClient", client)
	}
	if got := client.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UK8S client timeout = %s, want %s", got, 60*time.Second)
	}
	if got := client.GetConfig().Region; got != "cn-bj2" {
		t.Fatalf("UK8S client region = %q, want cn-bj2", got)
	}
	if config.Timeout != time.Second {
		t.Fatalf("newClient changed caller config timeout to %s", config.Timeout)
	}
}

func TestClientUsesProductRuntime(t *testing.T) {
	runtime := &testRuntime{config: &ucloud.Config{Region: "cn-bj2"}}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("clientFromMeta() error = %v", err)
	}
	if runtime.name != Name {
		t.Fatalf("ProductClient() name = %q, want %q", runtime.name, Name)
	}
	if runtime.calls != 1 {
		t.Fatalf("ProductClient() calls = %d, want 1", runtime.calls)
	}
	if client == nil {
		t.Fatal("clientFromMeta() returned a nil UK8S client")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("clientFromMeta accepted a non-runtime value")
	}
}

func TestUK8SServiceEmptyIdentifiersAreNotFound(t *testing.T) {
	client := newClient(&ucloud.Config{}, &auth.Credential{}, nil).(*sdkuk8s.UK8SClient)
	if _, err := describeUK8SClusterById(client, ""); err == nil || !isNotFoundError(err) {
		t.Fatalf("empty UK8S cluster ID error = %v, want provider not-found error", err)
	}
	if _, err := describeUK8SClusterNodeById(client, ""); err == nil || !isNotFoundError(err) {
		t.Fatalf("empty UK8S cluster node cluster ID error = %v, want provider not-found error", err)
	}
	if _, err := describeUK8SClusterNodeByResourceId(client, "cn-bj2", ""); err == nil || !isNotFoundError(err) {
		t.Fatalf("empty UK8S node ID error = %v, want provider not-found error", err)
	}
}

func TestUK8SValidationAndParsing(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"n-basic-2":        false,
		"o-standard-2":     false,
		"n-customized-2-6": false,
		"n-standard-3":     true,
		"invalid":          true,
	} {
		_, err := parseInstanceType(value)
		if got := err != nil; got != wantErr {
			t.Errorf("parseInstanceType(%q) error = %v, wantErr = %t", value, err, wantErr)
		}
	}

	parsed, err := parseInstanceType("n-basic-2")
	if err != nil {
		t.Fatalf("parseInstanceType(n-basic-2): %v", err)
	}
	if parsed.CPU != 2 || parsed.Memory != 4096 || parsed.HostType != "n" || parsed.HostScaleType != "basic" {
		t.Fatalf("parsed n-basic-2 = %#v", parsed)
	}

	for value, wantErr := range map[string]bool{
		"Password1!": true,
		"password":   false,
		"short1":     false,
	} {
		_, errors := validateInstancePassword(value, "password")
		if got := len(errors) == 0; got != wantErr {
			t.Errorf("validateInstancePassword(%q) errors = %v, valid = %t", value, errors, got)
		}
	}

	for value, wantErr := range map[string]bool{
		"uimage-abcde": true,
		"uimage-ab":    false,
		"image-abcde":  false,
	} {
		_, errors := validateUImageName(value, "image_id")
		if got := len(errors) == 0; got != wantErr {
			t.Errorf("validateUImageName(%q) errors = %v, valid = %t", value, errors, got)
		}
	}

	if upperCamelCvt.unconvert("dynamic") != "Dynamic" || upperCamelCvt.convert("Month") != "month" {
		t.Fatal("charge type conversion is incompatible with legacy behavior")
	}
	if upperCvt.unconvert("cloud_rssd") != "CLOUD_RSSD" || upperCvt.convert("CLOUD_SSD") != "cloud_ssd" {
		t.Fatal("disk type conversion is incompatible with legacy behavior")
	}
}

func TestUK8SCompatibilityHelpers(t *testing.T) {
	if got, want := timestampToString(0), time.Unix(0, 0).Format(time.RFC3339); got != want {
		t.Fatalf("timestampToString(0) = %q, want %q", got, want)
	}
	if !isNotFoundError(newNotFoundError("missing")) {
		t.Fatal("newNotFoundError was not recognized")
	}
	if isNotFoundError(fmt.Errorf("missing")) {
		t.Fatal("ordinary error was recognized as provider not-found error")
	}

	for value, wantErr := range map[int]bool{0: false, 9: false, 10: true} {
		_, errors := validateDuration(value, "duration")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("validateDuration(%d) errors = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[int]bool{20: false, 25: true, 1000: false, 1001: true} {
		_, errors := resourceUCloudUK8SCluster().Schema["master"].Elem.(*schema.Resource).Schema["data_disk_size"].ValidateFunc(value, "data_disk_size")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("cluster data_disk_size validation for %d = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

type testRuntime struct {
	config *ucloud.Config
	name   string
	calls  int
}

var _ product.RuntimeV1 = (*testRuntime)(nil)

func (runtime *testRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	if name != Name {
		return nil, fmt.Errorf("unexpected product %q", name)
	}
	runtime.name = name
	runtime.calls++
	return constructor(runtime.config, &auth.Credential{}, nil), nil
}
