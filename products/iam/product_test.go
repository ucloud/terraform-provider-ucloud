package iam

import (
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("iam"))); err != nil {
		t.Fatalf("register IAM: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate IAM provider: %v", err)
	}

	for _, name := range []string{
		"ucloud_iam_access_key",
		"ucloud_iam_user",
		"ucloud_iam_group",
		"ucloud_iam_group_membership",
		"ucloud_iam_project",
		"ucloud_iam_policy",
		"ucloud_iam_user_policy_attachment",
		"ucloud_iam_group_policy_attachment",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{
		"ucloud_iam_users",
		"ucloud_iam_groups",
		"ucloud_iam_projects",
		"ucloud_iam_policy",
		"ucloud_iam_policy_document",
	} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceSchemasKeepLegacyContract(t *testing.T) {
	resources := New().Registration().Resources
	wantFields := map[string][]string{
		"ucloud_iam_access_key":              {"user_name", "secret_file", "status", "secret", "pgp_key", "key_fingerprint", "encrypted_secret"},
		"ucloud_iam_user":                    {"name", "display_name", "email", "is_frozen", "login_enable", "status"},
		"ucloud_iam_group":                   {"name", "comment"},
		"ucloud_iam_group_membership":        {"group_name", "user_names"},
		"ucloud_iam_project":                 {"name", "create_time"},
		"ucloud_iam_policy":                  {"name", "comment", "policy", "scope", "create_time", "urn"},
		"ucloud_iam_user_policy_attachment":  {"user_name", "policy_urn", "project_id", "create_time"},
		"ucloud_iam_group_policy_attachment": {"group_name", "policy_urn", "project_id", "create_time"},
	}
	for name, fields := range wantFields {
		resource := resources[name]
		if resource == nil {
			t.Fatalf("resource %q is missing", name)
		}
		if len(resource.Schema) != len(fields) {
			t.Errorf("resource %q has %d fields, want %d", name, len(resource.Schema), len(fields))
		}
		for _, field := range fields {
			if resource.Schema[field] == nil {
				t.Errorf("resource %q is missing field %q", name, field)
			}
		}
		if resource.Create == nil || resource.Read == nil || resource.Delete == nil {
			t.Errorf("resource %q is missing CRUD callbacks", name)
		}
		if resource.CustomizeDiff != nil || resource.MigrateState != nil || resource.Timeouts != nil || len(resource.StateUpgraders) != 0 {
			t.Errorf("resource %q gained unsupported migration or lifecycle hooks", name)
		}
	}

	assertField(t, resources["ucloud_iam_access_key"].Schema["user_name"], schema.TypeString, true, false, false, true, false)
	assertField(t, resources["ucloud_iam_access_key"].Schema["status"], schema.TypeString, false, true, false, false, false)
	if got := resources["ucloud_iam_access_key"].Schema["status"].Default; got != iamStatusActive {
		t.Errorf("access key status default = %#v, want %q", got, iamStatusActive)
	}
	assertField(t, resources["ucloud_iam_access_key"].Schema["secret"], schema.TypeString, false, false, true, false, true)
	if resources["ucloud_iam_access_key"].Importer != nil {
		t.Error("access key must retain the legacy absence of an importer")
	}

	assertField(t, resources["ucloud_iam_user"].Schema["name"], schema.TypeString, true, false, false, true, false)
	assertField(t, resources["ucloud_iam_user"].Schema["email"], schema.TypeString, false, true, true, false, false)
	assertField(t, resources["ucloud_iam_user"].Schema["status"], schema.TypeString, false, false, true, false, false)
	if resources["ucloud_iam_user"].Importer == nil || resources["ucloud_iam_user"].Importer.State == nil {
		t.Error("IAM user importer is missing")
	}
	if got := resources["ucloud_iam_user"].Schema["login_enable"].Default; got != true {
		t.Errorf("login_enable default = %#v, want true", got)
	}

	assertField(t, resources["ucloud_iam_group"].Schema["name"], schema.TypeString, true, false, false, true, false)
	assertField(t, resources["ucloud_iam_group_membership"].Schema["group_name"], schema.TypeString, true, false, false, true, false)
	assertField(t, resources["ucloud_iam_group_membership"].Schema["user_names"], schema.TypeSet, true, false, false, false, false)
	assertField(t, resources["ucloud_iam_project"].Schema["name"], schema.TypeString, true, false, false, false, false)
	assertField(t, resources["ucloud_iam_project"].Schema["create_time"], schema.TypeString, false, false, true, false, false)

	assertField(t, resources["ucloud_iam_policy"].Schema["name"], schema.TypeString, true, false, false, true, false)
	assertField(t, resources["ucloud_iam_policy"].Schema["policy"], schema.TypeString, true, false, false, false, false)
	assertField(t, resources["ucloud_iam_policy"].Schema["scope"], schema.TypeString, true, false, false, true, false)
	assertField(t, resources["ucloud_iam_policy"].Schema["urn"], schema.TypeString, false, false, true, false, false)
	assertField(t, resources["ucloud_iam_user_policy_attachment"].Schema["project_id"], schema.TypeString, false, true, false, true, false)
	assertField(t, resources["ucloud_iam_group_policy_attachment"].Schema["project_id"], schema.TypeString, false, true, false, true, false)
	for _, name := range []string{
		"ucloud_iam_user",
		"ucloud_iam_group",
		"ucloud_iam_group_membership",
		"ucloud_iam_project",
		"ucloud_iam_policy",
		"ucloud_iam_user_policy_attachment",
		"ucloud_iam_group_policy_attachment",
	} {
		if resources[name].Importer == nil || resources[name].Importer.State == nil {
			t.Errorf("resource %q importer is missing", name)
		}
	}
}

func TestDataSourceSchemasKeepLegacyContract(t *testing.T) {
	dataSources := New().Registration().DataSources
	wantFields := map[string][]string{
		"ucloud_iam_users":           {"name_regex", "group_name", "names", "users"},
		"ucloud_iam_groups":          {"name_regex", "names", "groups"},
		"ucloud_iam_projects":        {"name_regex", "output_file", "total_count", "projects"},
		"ucloud_iam_policy":          {"name", "type", "urn", "comment", "scope", "policy", "create_time"},
		"ucloud_iam_policy_document": {"version", "statement", "output_file", "json"},
	}
	for name, fields := range wantFields {
		dataSource := dataSources[name]
		if dataSource == nil {
			t.Fatalf("data source %q is missing", name)
		}
		if len(dataSource.Schema) != len(fields) {
			t.Errorf("data source %q has %d fields, want %d", name, len(dataSource.Schema), len(fields))
		}
		if dataSource.Read == nil {
			t.Errorf("data source %q is missing Read callback", name)
		}
		for _, field := range fields {
			if dataSource.Schema[field] == nil {
				t.Errorf("data source %q is missing field %q", name, field)
			}
		}
	}

	users := dataSources["ucloud_iam_users"].Schema["users"]
	if users.Type != schema.TypeList || !users.Computed {
		t.Fatal("users must remain a computed TypeList")
	}
	userElement, ok := users.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("users element type = %T, want *schema.Resource", users.Elem)
	}
	for _, field := range []string{"name", "display_name", "email", "status", "login_enable"} {
		if userElement.Schema[field] == nil || !userElement.Schema[field].Computed {
			t.Errorf("users nested field %q must be computed", field)
		}
	}

	projects := dataSources["ucloud_iam_projects"].Schema["projects"]
	if projects.Type != schema.TypeList || !projects.Computed {
		t.Fatal("projects must remain a computed TypeList")
	}
	projectElement, ok := projects.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("projects element type = %T, want *schema.Resource", projects.Elem)
	}
	for _, field := range []string{"id", "name", "user_count", "create_time"} {
		if projectElement.Schema[field] == nil || !projectElement.Schema[field].Computed {
			t.Errorf("projects nested field %q must be computed", field)
		}
	}

	statement := dataSources["ucloud_iam_policy_document"].Schema["statement"]
	if statement.Type != schema.TypeList || !statement.Optional {
		t.Fatal("policy document statement must remain an optional TypeList")
	}
	statementElement, ok := statement.Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("statement element type = %T, want *schema.Resource", statement.Elem)
	}
	if statementElement.Schema["action"] == nil || !statementElement.Schema["action"].Required {
		t.Fatal("policy document action must remain required")
	}
}

func TestIAMClientUsesProductRuntime(t *testing.T) {
	runtime := &runtimeStub{}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}
	if runtime.name != Name || runtime.calls != 1 {
		t.Fatalf("runtime call = (%q, %d), want (%q, 1)", runtime.name, runtime.calls, Name)
	}
	if client == nil || client.iamconn == nil {
		t.Fatal("product client did not initialize the IAM SDK client")
	}
}

func TestIAMClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}

func TestIAMClientFromMetaRejectsUnexpectedProductClient(t *testing.T) {
	runtime := &runtimeStub{returnValue: &iam.IAMClient{}}
	if _, err := clientFromMeta(runtime); err == nil {
		t.Fatal("expected unexpected client type error")
	}
}

func TestAttachmentIDsRoundTrip(t *testing.T) {
	for _, test := range []struct {
		name      string
		policyURN string
		projectID string
	}{
		{name: "alice", policyURN: "urn:ucloud:iam:policy:read", projectID: ""},
		{name: "alice", policyURN: "urn:ucloud:iam:policy:write", projectID: "project-1"},
	} {
		userID := buildUCloudIAMUserPolicyAttachmentID(test.name, test.policyURN, test.projectID)
		userName, userPolicy, userProject, err := extractUCloudIAMUserPolicyAttachmentID(userID)
		if err != nil || userName != test.name || userPolicy != test.policyURN || userProject != test.projectID {
			t.Errorf("user attachment round trip = (%q, %q, %q, %v)", userName, userPolicy, userProject, err)
		}

		groupID := buildUCloudIAMGroupPolicyAttachmentID("admins", test.policyURN, test.projectID)
		groupName, groupPolicy, groupProject, err := extractUCloudIAMGroupPolicyAttachmentID(groupID)
		if err != nil || groupName != "admins" || groupPolicy != test.policyURN || groupProject != test.projectID {
			t.Errorf("group attachment round trip = (%q, %q, %q, %v)", groupName, groupPolicy, groupProject, err)
		}
	}
}

func TestAttachmentIDsRejectMalformedValues(t *testing.T) {
	for _, id := range []string{"", "account/", "account/user", "project/only/two", "unknown/user/policy"} {
		if _, _, _, err := extractUCloudIAMUserPolicyAttachmentID(id); err == nil {
			t.Errorf("malformed user attachment ID %q was accepted", id)
		}
		if _, _, _, err := extractUCloudIAMGroupPolicyAttachmentID(id); err == nil {
			t.Errorf("malformed group attachment ID %q was accepted", id)
		}
	}
}

func TestPolicyDocumentAssembly(t *testing.T) {
	got, err := assembleDataSourcePolicyJSON([]interface{}{
		map[string]interface{}{
			"effect":   "Allow",
			"action":   []interface{}{"GetObject", "ListObject"},
			"resource": []interface{}{"*"},
		},
	}, "1")
	if err != nil {
		t.Fatalf("assemble policy document: %v", err)
	}
	want := `{"Version":"1","Statement":[{"Action":["GetObject","ListObject"],"Effect":"Allow","Resource":["*"]}]}`
	if got != want {
		t.Fatalf("assembled policy document = %s, want %s", got, want)
	}

	resource := dataSourceUCloudIAMPolicyDocument()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"version": "1",
		"statement": []interface{}{
			map[string]interface{}{
				"effect":   "Deny",
				"action":   []interface{}{"DeleteObject"},
				"resource": []interface{}{"bucket/*"},
			},
		},
	})
	if err := resource.Read(data, nil); err != nil {
		t.Fatalf("read policy document data source: %v", err)
	}
	if got := data.Get("json").(string); !strings.Contains(got, `"Effect":"Deny"`) {
		t.Fatalf("policy document data source JSON = %s", got)
	}
	if data.Id() == "" {
		t.Fatal("policy document data source did not set an ID")
	}
}

func TestIAMCompatibilityHelpers(t *testing.T) {
	if got, want := timestampToString(0), time.Unix(0, 0).Format(time.RFC3339); got != want {
		t.Fatalf("timestampToString(0) = %q", got)
	}
	if got := getNotFoundMessage("iam", "missing"); got != "the specified iam missing is not found" {
		t.Fatalf("not found message = %q", got)
	}
	if !isNotFoundError(newNotFoundError("missing")) {
		t.Fatal("newNotFoundError was not recognized")
	}
	if got := (&ProviderError{errorCode: NotFound, message: "missing"}).Error(); !strings.Contains(got, "Code: Notfound") {
		t.Fatalf("provider error = %q", got)
	}
}

func assertField(t *testing.T, field *schema.Schema, typ schema.ValueType, required, optional, computed, forceNew, sensitive bool) {
	t.Helper()
	if field == nil {
		t.Fatal("schema field is nil")
	}
	if field.Type != typ || field.Required != required || field.Optional != optional || field.Computed != computed || field.ForceNew != forceNew || field.Sensitive != sensitive {
		t.Errorf("field flags = type=%v required=%t optional=%t computed=%t force_new=%t sensitive=%t; want type=%v required=%t optional=%t computed=%t force_new=%t sensitive=%t", field.Type, field.Required, field.Optional, field.Computed, field.ForceNew, field.Sensitive, typ, required, optional, computed, forceNew, sensitive)
	}
}

type runtimeStub struct {
	name        string
	calls       int
	returnValue interface{}
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	stub.name = name
	stub.calls++
	if stub.returnValue != nil {
		return stub.returnValue, nil
	}
	config := ucloud.NewConfig()
	return constructor(&config, &auth.Credential{}, nil), nil
}
