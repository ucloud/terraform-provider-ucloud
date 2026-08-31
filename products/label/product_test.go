package label

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("label", "labels"))); err != nil {
		t.Fatalf("register label: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate label provider: %v", err)
	}
	for _, name := range []string{"ucloud_label", "ucloud_label_attachment"} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_labels", "ucloud_label_resources"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceSchemasKeepLegacyContract(t *testing.T) {
	resources := New().Registration().Resources
	for name, fields := range map[string]map[string]schema.ValueType{
		"ucloud_label": {
			"key":   schema.TypeString,
			"value": schema.TypeString,
		},
		"ucloud_label_attachment": {
			"key":      schema.TypeString,
			"value":    schema.TypeString,
			"resource": schema.TypeString,
		},
	} {
		resource := resources[name]
		if resource == nil {
			t.Fatalf("resource %q is missing", name)
		}
		if len(resource.Schema) != len(fields) {
			t.Fatalf("resource %q has %d fields, want %d", name, len(resource.Schema), len(fields))
		}
		if resource.Create == nil || resource.Read == nil || resource.Delete == nil || resource.Update != nil {
			t.Fatalf("resource %q CRUD functions do not match legacy contract", name)
		}
		if resource.Importer == nil || resource.Importer.State == nil {
			t.Fatalf("resource %q importer is not configured", name)
		}
		if resource.CustomizeDiff != nil || resource.MigrateState != nil || resource.Timeouts != nil || len(resource.StateUpgraders) != 0 {
			t.Fatalf("resource %q gained unsupported migration or lifecycle hooks", name)
		}
		for fieldName, fieldType := range fields {
			field := resource.Schema[fieldName]
			if field == nil {
				t.Errorf("resource %q is missing field %q", name, fieldName)
				continue
			}
			if field.Type != fieldType || !field.Required || !field.ForceNew || field.Optional || field.Computed {
				t.Errorf("resource %q field %q flags = type=%v required=%t optional=%t computed=%t force_new=%t", name, fieldName, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew)
			}
		}
	}
}

func TestDataSourceSchemasKeepLegacyContract(t *testing.T) {
	dataSources := New().Registration().DataSources
	for name, fields := range map[string][]string{
		"ucloud_labels":          {"key_regex", "total_count", "output_file", "labels"},
		"ucloud_label_resources": {"key", "value", "resource_types", "project_ids", "output_file", "total_count", "resources"},
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

	labels := dataSources["ucloud_labels"].Schema["labels"]
	if labels.Type != schema.TypeList || !labels.Computed {
		t.Fatal("labels must remain a computed TypeList")
	}
	labelElement, ok := labels.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("labels element type = %T, want *schema.Resource", labels.Elem)
	}
	for _, field := range []string{"key", "value", "projects"} {
		if labelElement.Schema[field] == nil || !labelElement.Schema[field].Computed {
			t.Errorf("labels nested field %q must be computed", field)
		}
	}
	projects := labelElement.Schema["projects"]
	projectElement, ok := projects.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("projects element type = %T, want *schema.Resource", projects.Elem)
	}
	for _, field := range []string{"id", "name", "resource_types", "disabled_resource_types"} {
		if projectElement.Schema[field] == nil || !projectElement.Schema[field].Computed {
			t.Errorf("projects nested field %q must be computed", field)
		}
	}

	resources := dataSources["ucloud_label_resources"].Schema["resources"]
	if resources.Type != schema.TypeList || !resources.Computed {
		t.Fatal("resources must remain a computed TypeList")
	}
	resourceElement, ok := resources.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("resources element type = %T, want *schema.Resource", resources.Elem)
	}
	for _, field := range []string{"id", "name", "type"} {
		if resourceElement.Schema[field] == nil || !resourceElement.Schema[field].Computed {
			t.Errorf("resources nested field %q must be computed", field)
		}
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
	credential := auth.NewCredential()
	return constructor(&config, &credential, nil), nil
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
		t.Fatal("clientFromMeta returned a nil Label client")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}

func TestNewClientUsesLabelSDK(t *testing.T) {
	config := ucloud.NewConfig()
	credential := auth.NewCredential()
	client, ok := newClient(&config, &credential, nil).(*labelapi.LabelClient)
	if !ok {
		t.Fatalf("newClient returned %T, want *label.LabelClient", newClient(&config, &credential, nil))
	}
	if client.GetConfig() != &config {
		t.Fatal("label client did not retain the configured SDK client")
	}
}

func TestLegacyIDsAndHelpers(t *testing.T) {
	if got := buildUCloudLabelID("team", "blue"); got != "team#blue" {
		t.Fatalf("label ID = %q, want team#blue", got)
	}
	key, value, err := parseUCloudLabelID("team#blue")
	if err != nil || key != "team" || value != "blue" {
		t.Fatalf("parse label ID = (%q, %q, %v)", key, value, err)
	}
	if _, _, err := parseUCloudLabelID("invalid"); err == nil {
		t.Fatal("invalid label ID was accepted")
	}

	if got := buildUCloudLabelAttachmentID("team", "blue", "resource-1"); got != "team#blue#resource-1" {
		t.Fatalf("attachment ID = %q, want team#blue#resource-1", got)
	}
	key, value, resourceID, err := parseUCloudLabelAttachmentID("team#blue#resource-1")
	if err != nil || key != "team" || value != "blue" || resourceID != "resource-1" {
		t.Fatalf("parse attachment ID = (%q, %q, %q, %v)", key, value, resourceID, err)
	}
	if _, _, _, err := parseUCloudLabelAttachmentID("invalid"); err == nil {
		t.Fatal("invalid label attachment ID was accepted")
	}

	if got := interfaceSliceToStringSlice([]interface{}{"a", "b"}); strings.Join(got, ",") != "a,b" {
		t.Fatalf("interface slice conversion = %#v", got)
	}
	if !isNotFoundError(newNotFoundError("missing")) {
		t.Fatal("newNotFoundError was not recognized")
	}
	if isNotFoundError(os.ErrNotExist) {
		t.Fatal("os.ErrNotExist was unexpectedly recognized as provider not-found")
	}
}

func TestLegacyStateIsRetained(t *testing.T) {
	tests := []struct {
		name     string
		resource *schema.Resource
		state    *terraform.InstanceState
	}{
		{
			name:     "label",
			resource: resourceUCloudLabel(),
			state:    &terraform.InstanceState{ID: "team#blue", Attributes: map[string]string{"key": "team", "value": "blue"}},
		},
		{
			name:     "attachment",
			resource: resourceUCloudLabelAttachment(),
			state:    &terraform.InstanceState{ID: "team#blue#resource-1", Attributes: map[string]string{"key": "team", "value": "blue", "resource": "resource-1"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := test.resource.Data(test.state).State()
			if state == nil || state.ID != test.state.ID {
				t.Fatalf("legacy state = %#v, want ID %q", state, test.state.ID)
			}
			for key, value := range test.state.Attributes {
				if state.Attributes[key] != value {
					t.Errorf("state attribute %q = %q, want %q", key, state.Attributes[key], value)
				}
			}
		})
	}
}

func TestWriteToFileCompatibility(t *testing.T) {
	path := filepath.Join(t.TempDir(), "labels.json")
	if err := writeToFile(path, []map[string]interface{}{{"key": "team", "value": "blue"}}); err != nil {
		t.Fatalf("writeToFile() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(content), `"key": "team"`) || !strings.Contains(string(content), `"value": "blue"`) {
		t.Fatalf("output file = %s", content)
	}
}
