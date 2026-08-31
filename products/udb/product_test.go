package udb

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
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
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("db"))); err != nil {
		t.Fatalf("register UDB product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UDB product: %v", err)
	}

	if provider.ResourcesMap["ucloud_db_instance"] == nil {
		t.Fatal("ucloud_db_instance is not registered")
	}
	for _, name := range []string{"ucloud_db_instances", "ucloud_db_backups", "ucloud_db_parameter_groups"} {
		if provider.DataSourcesMap[name] == nil {
			t.Fatalf("%s is not registered", name)
		}
	}
}

func TestUDBClientPreservesLegacyHTTPTimeout(t *testing.T) {
	config := ucloud.NewConfig()
	credential := auth.NewCredential()
	rawClient := newClient(&config, &credential, nil)
	client, ok := rawClient.(*udb.UDBClient)
	if !ok {
		t.Fatalf("newClient() returned %T, want *udb.UDBClient", rawClient)
	}
	if got := client.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UDB client timeout = %s, want %s", got, 60*time.Second)
	}
}

func TestUDBClientUsesProductRuntime(t *testing.T) {
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
		t.Fatal("clientFromMeta() returned a nil UDB client")
	}
}

func TestDBInstanceResourceCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_db_instance"]
	if resource.Importer == nil || resource.Importer.State == nil {
		t.Fatal("ucloud_db_instance importer is not configured")
	}
	if resource.CustomizeDiff == nil {
		t.Fatal("ucloud_db_instance CustomizeDiff is not configured")
	}
	if resource.Timeouts == nil {
		t.Fatal("ucloud_db_instance timeouts are not configured")
	}
	for name, want := range map[string]time.Duration{
		"create": 30 * time.Minute,
		"update": 20 * time.Minute,
		"delete": 10 * time.Minute,
	} {
		var timeout *time.Duration
		switch name {
		case "create":
			timeout = resource.Timeouts.Create
		case "update":
			timeout = resource.Timeouts.Update
		case "delete":
			timeout = resource.Timeouts.Delete
		}
		if timeout == nil || *timeout != want {
			t.Fatalf("%s timeout = %v, want %s", name, timeout, want)
		}
	}

	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"availability_zone":         {typeValue: schema.TypeString, required: true, forceNew: true},
		"standby_zone":              {typeValue: schema.TypeString, optional: true, forceNew: true},
		"parameter_group":           {typeValue: schema.TypeString, optional: true, computed: true},
		"password":                  {typeValue: schema.TypeString, optional: true, computed: true, sensitive: true},
		"engine":                    {typeValue: schema.TypeString, required: true, forceNew: true},
		"engine_version":            {typeValue: schema.TypeString, required: true, forceNew: true},
		"name":                      {typeValue: schema.TypeString, optional: true, computed: true},
		"instance_storage":          {typeValue: schema.TypeInt, required: true},
		"instance_type":             {typeValue: schema.TypeString, required: true},
		"port":                      {typeValue: schema.TypeInt, optional: true, computed: true},
		"charge_type":               {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"duration":                  {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"vpc_id":                    {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"subnet_id":                 {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"from_backup_id":            {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"backup_count":              {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"backup_begin_time":         {typeValue: schema.TypeInt, optional: true, computed: true},
		"backup_date":               {typeValue: schema.TypeString, optional: true, computed: true},
		"backup_black_list":         {typeValue: schema.TypeSet, optional: true, computed: true},
		"tag":                       {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"allow_stopping_for_update": {typeValue: schema.TypeBool, optional: true},
		"private_ip":                {typeValue: schema.TypeString, computed: true},
		"status":                    {typeValue: schema.TypeString, computed: true},
		"create_time":               {typeValue: schema.TypeString, computed: true},
		"expire_time":               {typeValue: schema.TypeString, computed: true},
		"modify_time":               {typeValue: schema.TypeString, computed: true},
	}
	if len(resource.Schema) != len(wantFields) {
		t.Fatalf("ucloud_db_instance schema has %d fields, want %d", len(resource.Schema), len(wantFields))
	}
	for name, want := range wantFields {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("ucloud_db_instance schema is missing %q", name)
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Optional != want.optional || field.Computed != want.computed || field.ForceNew != want.forceNew || field.Sensitive != want.sensitive {
			t.Errorf("field %q flags = type=%v required=%t optional=%t computed=%t force_new=%t sensitive=%t", name, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew, field.Sensitive)
		}
	}
	if resource.Schema["backup_count"].Default != 7 {
		t.Fatalf("backup_count default = %#v, want 7", resource.Schema["backup_count"].Default)
	}
}

func TestDataSourceRegistrationCompatibility(t *testing.T) {
	registrations := New().Registration().DataSources
	wantFields := map[string][]string{
		"ucloud_db_instances": {
			"availability_zone", "ids", "name_regex", "output_file", "total_count", "db_instances",
		},
		"ucloud_db_backups": {
			"availability_zone", "project_id", "name_regex", "output_file", "total_count", "db_backups",
		},
		"ucloud_db_parameter_groups": {
			"availability_zone", "multi_az", "name_regex", "class_type", "output_file", "total_count", "parameter_groups",
		},
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

	instances := registrations["ucloud_db_instances"].Schema["db_instances"]
	if instances.Type != schema.TypeList || !instances.Computed {
		t.Fatal("db_instances must remain a computed TypeList")
	}
	instanceElement, ok := instances.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("db_instances element type = %T, want *schema.Resource", instances.Elem)
	}
	for _, field := range []string{"id", "availability_zone", "instance_type", "standby_zone", "name", "vpc_id", "subnet_id", "engine", "engine_version", "port", "private_ip", "instance_storage", "charge_type", "backup_count", "backup_date", "backup_begin_time", "backup_black_list", "tag", "status", "create_time", "expire_time", "modify_time"} {
		if instanceElement.Schema[field] == nil || !instanceElement.Schema[field].Computed {
			t.Errorf("db_instances nested field %q must be computed", field)
		}
	}

	backups := registrations["ucloud_db_backups"].Schema["db_backups"]
	if backups.Type != schema.TypeList || !backups.Computed {
		t.Fatal("db_backups must remain a computed TypeList")
	}
	parameterGroups := registrations["ucloud_db_parameter_groups"].Schema["parameter_groups"]
	if parameterGroups.Type != schema.TypeList || !parameterGroups.Computed {
		t.Fatal("parameter_groups must remain a computed TypeList")
	}
}

func TestUDBValidationCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_db_instance"]
	for value, wantErr := range map[string]bool{
		"mysql-ha-1":      false,
		"mysql-ha-nvme-2": false,
		"postgresql-ha-2": false,
		"mysql-basic-1":   true,
		"other-ha-1":      true,
		"mysql-ha-nvme-x": true,
	} {
		_, errors := resource.Schema["instance_type"].ValidateFunc(value, "instance_type")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("instance_type validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"2018_UClou":  false,
		"password123": true,
		"short":       true,
	} {
		_, errors := resource.Schema["password"].ValidateFunc(value, "password")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("password validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	for value, wantErr := range map[string]bool{
		"test.%":       false,
		"dbname.table": false,
		"invalid":      true,
		"db.%extra":    true,
	} {
		_, errors := resource.Schema["backup_black_list"].Elem.(*schema.Schema).ValidateFunc(value, "backup_black_list")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("backup_black_list validation errors for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}
