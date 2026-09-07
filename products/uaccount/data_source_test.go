package uaccount

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func TestProjectsReadPreservesLegacyBehavior(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCount++
		if request.Method != http.MethodPost {
			t.Errorf("request method = %q, want POST", request.Method)
		}
		if err := request.ParseForm(); err != nil {
			t.Errorf("parse request form: %v", err)
		}
		if got := request.Form.Get("Action"); got != "GetProjectList" {
			t.Errorf("request action = %q, want GetProjectList", got)
		}
		if got := request.Form.Get("Region"); got != "cn-bj" {
			t.Errorf("request region = %q, want cn-bj", got)
		}
		if got := request.Form.Get("ProjectId"); got != "project-1" {
			t.Errorf("request project ID = %q, want project-1", got)
		}
		if got := request.Form.Get("IsFinance"); got != "yes" {
			t.Errorf("request IsFinance = %q, want yes", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{
  "RetCode": 0,
  "ProjectCount": 2,
  "ProjectSet": [
    {
      "ProjectId": "project-1",
      "ProjectName": "keep-one",
      "ParentId": "parent-1",
      "ParentName": "parent",
      "ResourceCount": 7,
      "MemberCount": 3,
      "CreateTime": 1700000000
    },
    {
      "ProjectId": "project-2",
      "ProjectName": "drop-me",
      "ParentId": "parent-2",
      "ParentName": "parent",
      "ResourceCount": 8,
      "MemberCount": 4,
      "CreateTime": 1700000001
    }
  ]
}`)
	}))
	defer server.Close()

	outputFile := filepath.Join(t.TempDir(), "projects.json")
	data := schema.TestResourceDataRaw(t, dataSourceUCloudProjects().Schema, map[string]interface{}{
		"is_finance":  true,
		"name_regex":  "^keep",
		"output_file": outputFile,
	})
	runtime := newRuntimeForServer(server.URL, "cn-bj", "project-1")

	if err := dataSourceUCloudProjectsRead(data, runtime); err != nil {
		t.Fatalf("dataSourceUCloudProjectsRead() error = %v", err)
	}
	if runtime.calls != 1 || requestCount != 1 {
		t.Fatalf("runtime/API calls = (%d, %d), want (1, 1)", runtime.calls, requestCount)
	}
	if got := data.Id(); got != hashStringArray([]string{"project-1"}) {
		t.Fatalf("data source ID = %q, want hash of filtered project", got)
	}
	if got := data.Get("total_count").(int); got != 1 {
		t.Fatalf("total_count = %d, want 1", got)
	}
	projects := data.Get("projects").([]interface{})
	if len(projects) != 1 {
		t.Fatalf("projects length = %d, want 1", len(projects))
	}
	project := projects[0].(map[string]interface{})
	for name, want := range map[string]interface{}{
		"id":             "project-1",
		"name":           "keep-one",
		"parent_id":      "parent-1",
		"parent_name":    "parent",
		"resource_count": 7,
		"member_count":   3,
		"create_time":    time.Unix(1700000000, 0).Format(time.RFC3339),
	} {
		if got := project[name]; got != want {
			t.Errorf("projects.0.%s = %#v, want %#v", name, got, want)
		}
	}
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read project output: %v", err)
	}
	if !strings.Contains(string(content), `"id": "project-1"`) || strings.Contains(string(content), "drop-me") {
		t.Fatalf("project output has unexpected content: %s", content)
	}
}

func TestProjectsReadOmitsLegacyZeroValueIsFinance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := request.ParseForm(); err != nil {
			t.Errorf("parse request form: %v", err)
		}
		if _, ok := request.PostForm["IsFinance"]; ok {
			t.Errorf("request unexpectedly contains zero-value IsFinance: %#v", request.PostForm["IsFinance"])
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{"RetCode":0,"ProjectSet":[]}`)
	}))
	defer server.Close()

	data := schema.TestResourceDataRaw(t, dataSourceUCloudProjects().Schema, map[string]interface{}{
		"is_finance": false,
	})
	if err := dataSourceUCloudProjectsRead(data, newRuntimeForServer(server.URL, "cn-bj", "project-1")); err != nil {
		t.Fatalf("dataSourceUCloudProjectsRead() error = %v", err)
	}
	if got := data.Get("total_count").(int); got != 0 {
		t.Fatalf("total_count = %d, want 0", got)
	}
}

func TestZonesReadPreservesRegionFilteringAndOutput(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCount++
		if err := request.ParseForm(); err != nil {
			t.Errorf("parse request form: %v", err)
		}
		if got := request.Form.Get("Action"); got != "GetRegion" {
			t.Errorf("request action = %q, want GetRegion", got)
		}
		if got := request.Form.Get("Region"); got != "cn-sh" {
			t.Errorf("request region = %q, want cn-sh", got)
		}
		if got := request.Form.Get("ProjectId"); got != "project-2" {
			t.Errorf("request project ID = %q, want project-2", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, `{
  "RetCode": 0,
  "Regions": [
    {"Region": "cn-bj", "Zone": "cn-bj-01"},
    {"Region": "cn-sh", "Zone": "cn-sh-01"},
    {"Region": "cn-sh", "Zone": "cn-sh-02"}
  ]
}`)
	}))
	defer server.Close()

	outputFile := filepath.Join(t.TempDir(), "zones.json")
	data := schema.TestResourceDataRaw(t, dataSourceUCloudZones().Schema, map[string]interface{}{
		"output_file": outputFile,
	})
	runtime := newRuntimeForServer(server.URL, "cn-sh", "project-2")

	if err := dataSourceUCloudZonesRead(data, runtime); err != nil {
		t.Fatalf("dataSourceUCloudZonesRead() error = %v", err)
	}
	if runtime.calls != 1 || requestCount != 1 {
		t.Fatalf("runtime/API calls = (%d, %d), want (1, 1)", runtime.calls, requestCount)
	}
	if got := data.Id(); got != hashStringArray([]string{"cn-sh-01", "cn-sh-02"}) {
		t.Fatalf("data source ID = %q, want hash of filtered zones", got)
	}
	if got := data.Get("total_count").(int); got != 2 {
		t.Fatalf("total_count = %d, want 2", got)
	}
	zones := data.Get("zones").([]interface{})
	if len(zones) != 2 {
		t.Fatalf("zones length = %d, want 2", len(zones))
	}
	for index, want := range []string{"cn-sh-01", "cn-sh-02"} {
		zone := zones[index].(map[string]interface{})
		if got := zone["id"]; got != want {
			t.Errorf("zones.%d.id = %#v, want %q", index, got, want)
		}
	}
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read zone output: %v", err)
	}
	if !strings.Contains(string(content), "cn-sh-01") || strings.Contains(string(content), "cn-bj-01") {
		t.Fatalf("zone output has unexpected content: %s", content)
	}
}

func TestDataSourceReadErrorsKeepLegacyWording(t *testing.T) {
	tests := []struct {
		name       string
		dataSource *schema.Resource
		read       func(*schema.ResourceData, interface{}) error
		message    string
		wantPrefix string
	}{
		{
			name:       "projects",
			dataSource: dataSourceUCloudProjects(),
			read:       dataSourceUCloudProjectsRead,
			message:    `{"RetCode":10001,"Message":"project request failed"}`,
			wantPrefix: "error on reading project list, ",
		},
		{
			name:       "zones",
			dataSource: dataSourceUCloudZones(),
			read:       dataSourceUCloudZonesRead,
			message:    `{"RetCode":10002,"Message":"region request failed"}`,
			wantPrefix: "error on reading region list, ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(writer, test.message)
			}))
			defer server.Close()

			data := schema.TestResourceDataRaw(t, test.dataSource.Schema, map[string]interface{}{})
			err := test.read(data, newRuntimeForServer(server.URL, "cn-bj", "project-1"))
			if err == nil || !strings.HasPrefix(err.Error(), test.wantPrefix) {
				t.Fatalf("read error = %v, want prefix %q", err, test.wantPrefix)
			}
		})
	}
}

func newRuntimeForServer(serverURL, region, projectID string) *testRuntime {
	config := ucloud.NewConfig()
	config.BaseUrl = serverURL
	config.Region = region
	config.ProjectId = projectID
	return &testRuntime{config: config}
}
