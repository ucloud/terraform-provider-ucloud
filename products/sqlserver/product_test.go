package sqlserver

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
		t.Fatalf("register SQL Server product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with SQL Server product: %v", err)
	}

	if provider.ResourcesMap["ucloud_sqlserver_instance"] == nil {
		t.Fatal("ucloud_sqlserver_instance is not registered")
	}
	if provider.DataSourcesMap["ucloud_sqlserver_instances"] == nil {
		t.Fatal("ucloud_sqlserver_instances is not registered")
	}
}

func TestSQLServerClientPreservesTimeout(t *testing.T) {
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

func TestSQLServerClientUsesProductRuntime(t *testing.T) {
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

func TestSQLServerClientRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta("not-a-runtime"); err == nil {
		t.Fatal("clientFromMeta() accepted an invalid runtime")
	}
}

func TestSQLServerInstanceResourceCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_sqlserver_instance"]
	if resource.Importer == nil || resource.Importer.State == nil {
		t.Fatal("ucloud_sqlserver_instance importer is not configured")
	}
	if resource.Timeouts == nil || resource.Timeouts.Create == nil || resource.Timeouts.Delete == nil {
		t.Fatal("ucloud_sqlserver_instance timeouts are not configured")
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
		"cpu":               {typeValue: schema.TypeInt, required: true, forceNew: true},
		"memory":            {typeValue: schema.TypeInt, required: true, forceNew: true},
		"password":          {typeValue: schema.TypeString, optional: true, computed: true, sensitive: true, forceNew: true},
		"disk_space":        {typeValue: schema.TypeInt, required: true, forceNew: true},
		"port":              {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"charge_type":       {typeValue: schema.TypeString, optional: true, forceNew: true},
		"duration":          {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"vpc_id":            {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"subnet_id":         {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"tag":               {typeValue: schema.TypeString, optional: true, forceNew: true},
		"private_ip":        {typeValue: schema.TypeString, computed: true},
		"status":            {typeValue: schema.TypeString, computed: true},
		"create_time":       {typeValue: schema.TypeString, computed: true},
		"expire_time":       {typeValue: schema.TypeString, computed: true},
		"modify_time":       {typeValue: schema.TypeString, computed: true},
	}
	if len(resource.Schema) != len(wantFields) {
		t.Fatalf("ucloud_sqlserver_instance schema has %d fields, want %d", len(resource.Schema), len(wantFields))
	}
	for name, want := range wantFields {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("ucloud_sqlserver_instance schema is missing %q", name)
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Optional != want.optional || field.Computed != want.computed || field.ForceNew != want.forceNew || field.Sensitive != want.sensitive {
			t.Errorf("field %q flags = type=%v required=%t optional=%t computed=%t force_new=%t sensitive=%t", name, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew, field.Sensitive)
		}
	}
	if got := resource.Schema["db_version"].Default; got != "sqlserver-2019" {
		t.Fatalf("db_version default = %#v, want sqlserver-2019", got)
	}
	if got := resource.Schema["port"].Default; got != 1433 {
		t.Fatalf("port default = %#v, want 1433", got)
	}
}

func TestSQLServerValidationCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_sqlserver_instance"]
	for value, wantErr := range map[string]bool{
		"sqlserver-2017":   false,
		"sqlserver-2019":   false,
		"sqlserver-2022":   false,
		"sqlserver-2008R2": true,
	} {
		_, errors := resource.Schema["db_version"].ValidateFunc(value, "db_version")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("db_version validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[int]bool{
		2:  false,
		64: false,
		1:  true,
		65: true,
	} {
		_, errors := resource.Schema["cpu"].ValidateFunc(value, "cpu")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("cpu validation errors for %d = %v, wantErr = %t", value, errors, wantErr)
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
	dataSource := registrations["ucloud_sqlserver_instances"]
	if dataSource == nil {
		t.Fatal("data source ucloud_sqlserver_instances is missing")
	}
	for _, field := range []string{"availability_zone", "ids", "name_regex", "vpc_id", "output_file", "total_count", "sqlserver_instances"} {
		if dataSource.Schema[field] == nil {
			t.Errorf("data source ucloud_sqlserver_instances is missing field %q", field)
		}
	}
	list := dataSource.Schema["sqlserver_instances"]
	if list.Type != schema.TypeList || !list.Computed {
		t.Fatal("sqlserver_instances must remain a computed TypeList")
	}
}
