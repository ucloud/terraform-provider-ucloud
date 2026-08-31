package uads

import (
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	sdkuads "github.com/ucloud/ucloud-sdk-go/services/uads"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type testRuntime struct {
	name        string
	calls       int
	constructor product.ClientConstructor
}

var _ product.RuntimeV1 = (*testRuntime)(nil)

func (runtime *testRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	runtime.name = name
	runtime.calls++
	runtime.constructor = constructor
	config := ucloud.NewConfig()
	credential := auth.NewCredential()
	return constructor(&config, &credential, nil), nil
}

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("anti_ddos"))); err != nil {
		t.Fatalf("register UADS product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UADS product: %v", err)
	}

	for _, name := range []string{
		"ucloud_anti_ddos_instance",
		"ucloud_anti_ddos_allowed_domain",
		"ucloud_anti_ddos_ip",
		"ucloud_anti_ddos_rule",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_anti_ddos_instances", "ucloud_anti_ddos_ips"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceCompatibility(t *testing.T) {
	resources := New().Registration().Resources
	wantFields := map[string]map[string]fieldContract{
		"ucloud_anti_ddos_instance": {
			"name":               {schema.TypeString, true, false, false, false},
			"area":               {schema.TypeString, true, false, false, true},
			"data_center":        {schema.TypeString, true, false, false, true},
			"bandwidth":          {schema.TypeInt, true, false, false, false},
			"base_defence_value": {schema.TypeInt, true, false, false, false},
			"max_defence_value":  {schema.TypeInt, true, false, false, false},
			"charge_type":        {schema.TypeString, false, true, false, true},
			"duration":           {schema.TypeInt, false, true, false, true},
			"create_time":        {schema.TypeString, false, false, true, false},
			"expire_time":        {schema.TypeString, false, false, true, false},
			"status":             {schema.TypeString, false, false, true, false},
		},
		"ucloud_anti_ddos_allowed_domain": {
			"domain":      {schema.TypeString, true, false, false, true},
			"comment":     {schema.TypeString, false, true, true, false},
			"instance_id": {schema.TypeString, true, false, false, true},
			"status":      {schema.TypeString, false, false, true, false},
		},
		"ucloud_anti_ddos_ip": {
			"instance_id": {schema.TypeString, true, false, false, true},
			"comment":     {schema.TypeString, false, true, true, false},
			"status":      {schema.TypeString, false, false, true, false},
			"domain":      {schema.TypeString, false, false, true, false},
			"ip":          {schema.TypeString, false, false, true, false},
		},
		"ucloud_anti_ddos_rule": {
			"instance_id":           {schema.TypeString, true, false, false, true},
			"ip":                    {schema.TypeString, true, false, false, true},
			"port":                  {schema.TypeInt, false, true, false, true},
			"real_server_type":      {schema.TypeString, true, false, false, true},
			"real_servers":          {schema.TypeList, true, false, false, false},
			"toa_id":                {schema.TypeInt, false, true, false, false},
			"real_server_detection": {schema.TypeBool, false, true, false, false},
			"backup_server":         {schema.TypeMap, false, true, false, false},
			"comment":               {schema.TypeString, false, true, true, false},
			"status":                {schema.TypeString, false, false, true, false},
			"rule_index":            {schema.TypeInt, false, false, true, false},
			"rule_id":               {schema.TypeString, false, false, true, false},
		},
	}

	for name, fields := range wantFields {
		resource := resources[name]
		assertResource(t, resource, name)
		if len(resource.Schema) != len(fields) {
			t.Errorf("%s schema has %d fields, want %d", name, len(resource.Schema), len(fields))
		}
		for fieldName, want := range fields {
			field, ok := resource.Schema[fieldName]
			if !ok {
				t.Errorf("%s schema is missing %q", name, fieldName)
				continue
			}
			if got := fieldContractOf(field); got != want {
				t.Errorf("%s.%s = %#v, want %#v", name, fieldName, got, want)
			}
		}
	}

	instance := resources["ucloud_anti_ddos_instance"]
	if instance.Schema["duration"].ValidateFunc == nil || instance.Schema["name"].ValidateFunc == nil {
		t.Fatal("instance validation functions are missing")
	}
	if resources["ucloud_anti_ddos_rule"].Schema["toa_id"].Default != 200 {
		t.Fatalf("toa_id default = %#v, want 200", resources["ucloud_anti_ddos_rule"].Schema["toa_id"].Default)
	}
	if resources["ucloud_anti_ddos_rule"].Schema["real_server_detection"].Default != false {
		t.Fatalf("real_server_detection default = %#v, want false", resources["ucloud_anti_ddos_rule"].Schema["real_server_detection"].Default)
	}
}

func TestResourceCallbacksImportersCustomizeDiffAndTimeouts(t *testing.T) {
	for name, resource := range New().Registration().Resources {
		assertResource(t, resource, name)
		if resource.Importer == nil || resource.Importer.State == nil {
			t.Errorf("%s importer is missing", name)
		}
		if resource.CustomizeDiff == nil {
			t.Errorf("%s CustomizeDiff is missing", name)
		}
		if resource.Timeouts != nil {
			t.Errorf("%s unexpectedly gained resource timeouts", name)
		}
	}
}

func TestDataSourceCompatibility(t *testing.T) {
	dataSources := New().Registration().DataSources
	wantFields := map[string][]string{
		"ucloud_anti_ddos_instances": {
			"ids", "name_regex", "output_file", "total_count", "instances",
		},
		"ucloud_anti_ddos_ips": {
			"instance_id", "ips", "output_file", "total_count",
		},
	}
	for name, fields := range wantFields {
		dataSource := dataSources[name]
		if dataSource == nil || dataSource.Read == nil {
			t.Fatalf("data source %q is missing read callback", name)
		}
		for _, field := range fields {
			if dataSource.Schema[field] == nil {
				t.Errorf("data source %q is missing field %q", name, field)
			}
		}
	}

	instances := dataSources["ucloud_anti_ddos_instances"].Schema["instances"]
	if instances.Type != schema.TypeList || !instances.Computed {
		t.Fatal("instances must remain a computed TypeList")
	}
	instanceElement, ok := instances.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("instances element type = %T, want *schema.Resource", instances.Elem)
	}
	for _, field := range []string{
		"id", "name", "area", "data_center", "bandwidth", "base_defence_value",
		"max_defence_value", "charge_type", "create_time", "expire_time", "status",
	} {
		if instanceElement.Schema[field] == nil || !instanceElement.Schema[field].Computed {
			t.Errorf("instances nested field %q must be computed", field)
		}
	}

	ips := dataSources["ucloud_anti_ddos_ips"].Schema["ips"]
	if ips.Type != schema.TypeList || !ips.Computed {
		t.Fatal("ips must remain a computed TypeList")
	}
	ipElement, ok := ips.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ips element type = %T, want *schema.Resource", ips.Elem)
	}
	for _, field := range []string{"instance_id", "ip", "domain", "comment", "proxy_ips", "status"} {
		if ipElement.Schema[field] == nil || !ipElement.Schema[field].Computed {
			t.Errorf("ips nested field %q must be computed", field)
		}
	}
}

func TestClientUsesProductRuntimeLazily(t *testing.T) {
	runtime := &testRuntime{}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("clientFromMeta() error = %v", err)
	}
	if runtime.name != Name || runtime.calls != 1 {
		t.Fatalf("ProductClient() call = (%q, %d), want (%q, 1)", runtime.name, runtime.calls, Name)
	}
	if runtime.constructor == nil {
		t.Fatal("ProductClient() did not receive a constructor")
	}
	if client == nil {
		t.Fatal("clientFromMeta() returned a nil UADS client")
	}
	if got := client.GetConfig(); got == nil {
		t.Fatal("UADS client has no config")
	}
	if _, ok := interface{}(client).(*sdkuads.UADSClient); !ok {
		t.Fatalf("clientFromMeta() returned %T, want *uads.UADSClient", client)
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil || !strings.Contains(err.Error(), "invalid provider runtime") {
		t.Fatalf("clientFromMeta() error = %v, want invalid runtime error", err)
	}
}

func TestCompatibilityHelpersAndValidation(t *testing.T) {
	if got := upperCamelCvt.convert("Month"); got != "month" {
		t.Fatalf("upperCamelCvt.convert(Month) = %q, want month", got)
	}
	if got := upperCamelCvt.unconvert("month"); got != "Month" {
		t.Fatalf("upperCamelCvt.unconvert(month) = %q, want Month", got)
	}
	if got := uadsAllowedDomainStatusCvt.convert(uadsAllowedDomainStatusSuccess); got != "Success" {
		t.Fatalf("allowed domain status = %q, want Success", got)
	}
	if got, want := timestampToString(0), time.Unix(0, 0).Format(time.RFC3339); got != want {
		t.Fatalf("timestampToString(0) = %q", got)
	}
	if got := hashStringArray([]string{"uads-1"}); got == "" {
		t.Fatal("hashStringArray returned an empty state ID")
	}

	for value, wantErr := range map[string]bool{"uads-1": false, "": true} {
		_, errors := validateAntiDDoSInstanceName(value, "name")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("name validation for %q errors = %v, wantErr %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[int]bool{0: false, 254: false, 255: true} {
		_, errors := validateToaID(value, "toa_id")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("toa validation for %d errors = %v, wantErr %t", value, errors, wantErr)
		}
	}

	raw := map[string]interface{}{
		"base_defence_value": 50,
		"max_defence_value":  30,
	}
	data := schema.TestResourceDataRaw(t, resourceUCloudAntiDDoSInstance().Schema, raw)
	if err := validateAntiDDoSInstance(data); err == nil || err.Error() != "max_defence_value 30 cannot be less than base_defence_value 50" {
		t.Fatalf("validateAntiDDoSInstance() error = %v", err)
	}

	for _, tc := range []struct {
		input map[string]interface{}
		port  int
		err   bool
	}{
		{input: map[string]interface{}{"port": "8080"}, port: 8080},
		{input: map[string]interface{}{"port": 8080}, port: 8080},
		{input: map[string]interface{}{"port": "bad"}, err: true},
		{input: map[string]interface{}{}, port: 0},
	} {
		got, err := getAntiDDoSRuleBackupServerPort(tc.input)
		if got != tc.port || (err != nil) != tc.err {
			t.Errorf("backup port for %#v = (%d, %v), want (%d, %t)", tc.input, got, err, tc.port, tc.err)
		}
	}
}

func TestInstanceDataSourceSavePreservesStateShape(t *testing.T) {
	data := schema.TestResourceDataRaw(t, dataSourceUCloudAntiDDoSInstances().Schema, map[string]interface{}{})
	instances := []sdkuads.ServiceInfo{
		{
			ResourceId:             "uads-1",
			Name:                   "instance",
			AreaLine:               "EastChina",
			EngineRoom:             []string{"Zaozhuang"},
			SrcBandwidth:           50,
			DefenceDDosBaseFlowArr: []int{30},
			DefenceDDosMaxFlowArr:  []int{60},
			ChargeType:             "Month",
			CreateTime:             1,
			ExpiredTime:            2,
			DefenceStatus:          "Running",
		},
	}
	if err := dataSourceUCloudAntiDDoSInstancesSave(data, instances); err != nil {
		t.Fatalf("save UADS instances: %v", err)
	}
	if data.Id() != hashStringArray([]string{"uads-1"}) {
		t.Fatalf("data source ID = %q, want %q", data.Id(), hashStringArray([]string{"uads-1"}))
	}
	if got := data.Get("total_count").(int); got != 1 {
		t.Fatalf("total_count = %d, want 1", got)
	}
	items := data.Get("instances").([]interface{})
	if len(items) != 1 {
		t.Fatalf("instances length = %d, want 1", len(items))
	}
	item := items[0].(map[string]interface{})
	if item["id"] != "uads-1" || item["charge_type"] != "month" || item["data_center"] != "Zaozhuang" {
		t.Fatalf("saved instance state = %#v", item)
	}
}

type fieldContract struct {
	typeValue schema.ValueType
	required  bool
	optional  bool
	computed  bool
	forceNew  bool
}

func fieldContractOf(field *schema.Schema) fieldContract {
	return fieldContract{
		typeValue: field.Type,
		required:  field.Required,
		optional:  field.Optional,
		computed:  field.Computed,
		forceNew:  field.ForceNew,
	}
}

func assertResource(t *testing.T, resource *schema.Resource, name string) {
	t.Helper()
	if resource == nil {
		t.Fatalf("resource %q is nil", name)
	}
	if resource.Create == nil || resource.Read == nil || resource.Update == nil || resource.Delete == nil {
		t.Fatalf("resource %q CRUD callbacks are incomplete", name)
	}
}
