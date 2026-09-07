package uaccount

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type testRuntime struct {
	name        string
	calls       int
	constructor product.ClientConstructor
	config      ucloud.Config
	credential  auth.Credential
	returnValue interface{}
}

var _ product.RuntimeV1 = (*testRuntime)(nil)

func (runtime *testRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	runtime.name = name
	runtime.calls++
	runtime.constructor = constructor
	if runtime.returnValue != nil {
		return runtime.returnValue, nil
	}
	return constructor(&runtime.config, &runtime.credential, nil), nil
}

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("projects", "zones"))); err != nil {
		t.Fatalf("register UAccount product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UAccount product: %v", err)
	}

	if len(provider.ResourcesMap) != 0 {
		t.Fatalf("unexpected UAccount resources: %#v", provider.ResourcesMap)
	}
	for _, name := range []string{"ucloud_projects", "ucloud_zones"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestDataSourceSchemasKeepLegacyContract(t *testing.T) {
	dataSources := New().Registration().DataSources

	projects := dataSources["ucloud_projects"]
	assertDataSource(t, projects, "ucloud_projects")
	assertSchemaFields(t, projects, map[string]schemaFieldContract{
		"is_finance":  {typeValue: schema.TypeBool, optional: true},
		"name_regex":  {typeValue: schema.TypeString, optional: true},
		"output_file": {typeValue: schema.TypeString, optional: true},
		"total_count": {typeValue: schema.TypeInt, computed: true},
		"projects":    {typeValue: schema.TypeList, computed: true},
	})
	if projects.Schema["name_regex"].ValidateFunc == nil {
		t.Fatal("ucloud_projects.name_regex validation is missing")
	}
	if projects.Schema["is_finance"].Default != nil || projects.Schema["is_finance"].DefaultFunc != nil {
		t.Fatal("ucloud_projects.is_finance unexpectedly has a default")
	}
	projectItems, ok := projects.Schema["projects"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ucloud_projects.projects element type = %T, want *schema.Resource", projects.Schema["projects"].Elem)
	}
	for _, name := range []string{"id", "name", "parent_id", "parent_name", "resource_count", "member_count", "create_time"} {
		field := projectItems.Schema[name]
		if field == nil || !field.Computed {
			t.Errorf("ucloud_projects.projects nested field %q must be computed", name)
		}
	}

	zones := dataSources["ucloud_zones"]
	assertDataSource(t, zones, "ucloud_zones")
	assertSchemaFields(t, zones, map[string]schemaFieldContract{
		"output_file": {typeValue: schema.TypeString, optional: true},
		"total_count": {typeValue: schema.TypeInt, computed: true},
		"zones":       {typeValue: schema.TypeList, computed: true},
	})
	zoneItems, ok := zones.Schema["zones"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ucloud_zones.zones element type = %T, want *schema.Resource", zones.Schema["zones"].Elem)
	}
	if field := zoneItems.Schema["id"]; field == nil || !field.Computed || field.Type != schema.TypeString {
		t.Fatal("ucloud_zones.zones.id must be a computed string")
	}
}

func TestClientUsesProductRuntimeLazily(t *testing.T) {
	runtime := &testRuntime{config: ucloud.NewConfig()}
	runtime.config.Region = "cn-bj"
	runtime.config.ProjectId = "project-1"

	_ = New().Registration()
	if runtime.calls != 0 {
		t.Fatalf("registration constructed a client, calls = %d", runtime.calls)
	}

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
		t.Fatal("clientFromMeta() returned a nil UAccount client")
	}
	if got := client.GetConfig(); got == nil || got.Region != "cn-bj" || got.ProjectId != "project-1" {
		t.Fatalf("UAccount client config = %#v, want region/project ID preserved", got)
	}
	if got := client.GetMeta().Product; got != "UAccount" {
		t.Fatalf("UAccount client product metadata = %q, want UAccount", got)
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil || !strings.Contains(err.Error(), "invalid provider runtime") {
		t.Fatalf("clientFromMeta() error = %v, want invalid runtime error", err)
	}
}

func TestClientFromMetaRejectsUnexpectedProductClient(t *testing.T) {
	runtime := &testRuntime{returnValue: &ucloud.Client{}}
	if _, err := clientFromMeta(runtime); err == nil || !strings.Contains(err.Error(), "unexpected type") {
		t.Fatalf("clientFromMeta() error = %v, want unexpected client type error", err)
	}
}

func TestCompatibilityHelpers(t *testing.T) {
	if got := boolLowerCvt.convert(true); got != "yes" {
		t.Fatalf("boolLowerCvt.convert(true) = %q, want yes", got)
	}
	if got := boolLowerCvt.convert(false); got != "no" {
		t.Fatalf("boolLowerCvt.convert(false) = %q, want no", got)
	}
	if got := hashStringArray([]string{"project-1", "project-2"}); got == "" {
		t.Fatal("hashStringArray returned an empty ID")
	}
	if got := hashStringArray([]string{"project-1", "project-2"}); got == hashStringArray([]string{"project-2", "project-1"}) {
		t.Fatal("hashStringArray lost legacy ordering")
	}
	if got, want := timestampToString(0), time.Unix(0, 0).Format(time.RFC3339); got != want {
		t.Fatalf("timestampToString(0) = %q, want %q", got, want)
	}

	path := filepath.Join(t.TempDir(), "projects.json")
	if err := writeToFile(path, []map[string]interface{}{{"id": "project-1"}}); err != nil {
		t.Fatalf("writeToFile() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if got, want := string(content), "[\n\t{\n\t\t\"id\": \"project-1\"\n\t}\n]"; got != want {
		t.Fatalf("output file = %q, want %q", got, want)
	}
}

type schemaFieldContract struct {
	typeValue schema.ValueType
	required  bool
	optional  bool
	computed  bool
}

func assertDataSource(t *testing.T, dataSource *schema.Resource, name string) {
	t.Helper()
	if dataSource == nil {
		t.Fatalf("data source %q is missing", name)
	}
	if dataSource.Read == nil {
		t.Fatalf("data source %q is missing Read callback", name)
	}
}

func assertSchemaFields(t *testing.T, dataSource *schema.Resource, fields map[string]schemaFieldContract) {
	t.Helper()
	if len(dataSource.Schema) != len(fields) {
		t.Errorf("data source schema has %d fields, want %d", len(dataSource.Schema), len(fields))
	}
	for name, want := range fields {
		field := dataSource.Schema[name]
		if field == nil {
			t.Errorf("data source schema is missing %q", name)
			continue
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Optional != want.optional || field.Computed != want.computed {
			t.Errorf("field %q flags = type=%v required=%t optional=%t computed=%t, want type=%v required=%t optional=%t computed=%t", name, field.Type, field.Required, field.Optional, field.Computed, want.typeValue, want.required, want.optional, want.computed)
		}
	}
}
