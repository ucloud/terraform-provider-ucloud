package udb_mysql

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type testRuntime struct {
	name        string
	constructor product.ClientConstructor
}

func (runtime *testRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	runtime.name = name
	runtime.constructor = constructor
	config := ucloud.NewConfig()
	credential := auth.NewCredential()
	return constructor(&config, &credential, nil), nil
}

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New())); err != nil {
		t.Fatalf("register MySQL product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with MySQL product: %v", err)
	}

	if provider.ResourcesMap["ucloud_udb_mysql_instance"] == nil {
		t.Fatal("ucloud_udb_mysql_instance is not registered")
	}
	for _, name := range []string{"ucloud_udb_mysql_instances", "ucloud_udb_mysql_parameter_templates", "ucloud_udb_mysql_machine_types"} {
		if provider.DataSourcesMap[name] == nil {
			t.Fatalf("%s is not registered", name)
		}
	}
}

func TestMySQLClientPreservesTimeout(t *testing.T) {
	config := ucloud.NewConfig()
	credential := auth.NewCredential()
	rawClient := newClient(&config, &credential, nil)
	client, ok := rawClient.(*ucloud.Client)
	if !ok {
		t.Fatalf("newClient() returned %T, want *ucloud.Client", rawClient)
	}
	if got := client.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("client timeout = %s, want %s", got, 60*time.Second)
	}
}

func TestMySQLClientUsesProductRuntime(t *testing.T) {
	runtime := &testRuntime{}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("clientFromMeta() error = %v", err)
	}
	if runtime.name != Name {
		t.Fatalf("ProductClient() name = %q, want %q", runtime.name, Name)
	}
	if runtime.constructor == nil {
		t.Fatal("ProductClient() did not receive a client constructor")
	}
	if client == nil {
		t.Fatal("clientFromMeta() returned a nil client")
	}
}

func TestMySQLClientRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta("not-a-runtime"); err == nil {
		t.Fatal("clientFromMeta() accepted an invalid runtime")
	}
}

func TestMySQLInstanceResourceCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_udb_mysql_instance"]
	if resource.Importer == nil || resource.Importer.State == nil {
		t.Fatal("ucloud_udb_mysql_instance importer is not configured")
	}
	if resource.Update == nil {
		t.Fatal("ucloud_udb_mysql_instance update is not configured")
	}
	if resource.Timeouts == nil || resource.Timeouts.Create == nil || resource.Timeouts.Update == nil || resource.Timeouts.Delete == nil {
		t.Fatal("ucloud_udb_mysql_instance timeouts are not configured")
	}

	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"availability_zone": {typeValue: schema.TypeString, required: true, forceNew: true},
		"name":              {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"db_version":        {typeValue: schema.TypeString, optional: true, forceNew: true},
		"machine_type":      {typeValue: schema.TypeString, required: true},
		"password":          {typeValue: schema.TypeString, optional: true, computed: true, sensitive: true, forceNew: true},
		"instance_mode":     {typeValue: schema.TypeString, optional: true, forceNew: true},
		"disk_space":        {typeValue: schema.TypeInt, required: true},
		"port":              {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"charge_type":       {typeValue: schema.TypeString, optional: true, forceNew: true},
		"duration":          {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"param_group_id":    {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"vpc_id":            {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"subnet_id":         {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"tag":               {typeValue: schema.TypeString, optional: true, forceNew: true},
		"private_ip":        {typeValue: schema.TypeString, computed: true},
		"status":            {typeValue: schema.TypeString, computed: true},
		"cpu":               {typeValue: schema.TypeInt, computed: true},
		"memory":            {typeValue: schema.TypeInt, computed: true},
		"create_time":       {typeValue: schema.TypeString, computed: true},
		"expire_time":       {typeValue: schema.TypeString, computed: true},
		"modify_time":       {typeValue: schema.TypeString, computed: true},
	}
	if len(resource.Schema) != len(wantFields) {
		t.Fatalf("ucloud_udb_mysql_instance schema has %d fields, want %d", len(resource.Schema), len(wantFields))
	}
	if got := resource.Schema["instance_mode"].Default; got != "HA" {
		t.Fatalf("instance_mode default = %#v, want HA", got)
	}
	if got := resource.Schema["db_version"].Default; got != "mysql-8.0" {
		t.Fatalf("db_version default = %#v, want mysql-8.0", got)
	}
	for name, want := range wantFields {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("ucloud_udb_mysql_instance schema is missing %q", name)
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Optional != want.optional || field.Computed != want.computed || field.ForceNew != want.forceNew || field.Sensitive != want.sensitive {
			t.Errorf("field %q flags = type=%v required=%t optional=%t computed=%t force_new=%t sensitive=%t", name, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew, field.Sensitive)
		}
	}
}

func TestMySQLValidationCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_udb_mysql_instance"]
	for value, wantErr := range map[string]bool{
		"o.mysql2m.xlarge":   false,
		"o.mysql8m.16xlarge": false,
		"o.mysql9m.xlarge":   true,
		"n.mysql2m.small":    true,
	} {
		_, errors := resource.Schema["machine_type"].ValidateFunc(value, "machine_type")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("machine_type validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"mysql-5.7": false,
		"mysql-8.0": false,
		"mysql-8.4": false,
		"mysql-9.0": true,
	} {
		_, errors := resource.Schema["db_version"].ValidateFunc(value, "db_version")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("db_version validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"2018_UClou": false,
		"short":      true,
	} {
		_, errors := resource.Schema["password"].ValidateFunc(value, "password")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("password validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

func TestDataSourceRegistrationCompatibility(t *testing.T) {
	registrations := New().Registration().DataSources
	wantFields := map[string][]string{
		"ucloud_udb_mysql_instances":           {"availability_zone", "ids", "name_regex", "vpc_id", "output_file", "total_count", "udb_mysql_instances"},
		"ucloud_udb_mysql_parameter_templates": {"availability_zone", "db_version", "name_regex", "output_file", "total_count", "udb_mysql_parameter_templates"},
		"ucloud_udb_mysql_machine_types":       {"availability_zone", "instance_mode", "output_file", "total_count", "udb_mysql_machine_types"},
	}
	for name, fields := range wantFields {
		dataSource := registrations[name]
		if dataSource == nil {
			t.Fatalf("data source %q is missing", name)
		}
		for _, field := range fields {
			if dataSource.Schema[field] == nil {
				t.Errorf("data source %q is missing field %q", name, field)
			}
		}
	}

	for name, listField := range map[string]string{
		"ucloud_udb_mysql_instances":           "udb_mysql_instances",
		"ucloud_udb_mysql_parameter_templates": "udb_mysql_parameter_templates",
		"ucloud_udb_mysql_machine_types":       "udb_mysql_machine_types",
	} {
		list := registrations[name].Schema[listField]
		if list.Type != schema.TypeList || !list.Computed {
			t.Errorf("%s %s must remain a computed TypeList", name, listField)
		}
	}
}
