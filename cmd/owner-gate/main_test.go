package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMarksProductOwnerAsAutoMergeEligible(t *testing.T) {
	root, eventPath, configPath := ownerGateFixture(t, "ProductOwner")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasSuffix(request.URL.Path, "/pulls/42/files") {
			t.Fatalf("unexpected API request: %s", request.URL)
		}
		_, _ = io.WriteString(writer, `[{"filename":"products/us3/product.go","status":"modified"}]`)
	}))
	defer server.Close()

	var stdout, stderr strings.Builder
	outputs := filepath.Join(t.TempDir(), "github-output")
	env := testEnvironment(eventPath, server.URL, outputs)
	if err := run(root, []string{"-config", configPath}, &stdout, &stderr, env); err != nil {
		t.Fatalf("run() error = %v, stderr = %s", err, stderr.String())
	}
	var decision map[string]interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &decision); err != nil {
		t.Fatalf("decode decision: %v", err)
	}
	if decision["type"] != "product" || decision["autoMergeEligible"] != true || decision["blocking"] != false {
		t.Fatalf("decision = %#v", decision)
	}
	content, err := os.ReadFile(outputs)
	if err != nil {
		t.Fatalf("read GitHub outputs: %v", err)
	}
	if !strings.Contains(string(content), "autoMergeEligible=true") {
		t.Fatalf("GitHub outputs = %s", content)
	}
}

func TestRunClearsPlatformAfterCurrentAdminApproval(t *testing.T) {
	root, eventPath, configPath := ownerGateFixture(t, "OtherUser")
	head := "0123456789abcdef0123456789abcdef01234567"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case strings.HasSuffix(request.URL.Path, "/pulls/42/files"):
			_, _ = io.WriteString(writer, `[{"filename":"products/us3/product.go","status":"modified"}]`)
		case strings.HasSuffix(request.URL.Path, "/pulls/42/reviews"):
			_, _ = io.WriteString(writer, `[{"id":1,"state":"APPROVED","commit_id":"`+head+`","user":{"login":"AdminUser"}}]`)
		case strings.HasSuffix(request.URL.Path, "/collaborators/otheruser/permission"):
			_, _ = io.WriteString(writer, `{"permission":"pull"}`)
		case strings.HasSuffix(request.URL.Path, "/collaborators/adminuser/permission"):
			_, _ = io.WriteString(writer, `{"permission":"admin"}`)
		default:
			t.Fatalf("unexpected API request: %s", request.URL)
		}
	}))
	defer server.Close()

	payload, err := os.ReadFile(eventPath)
	if err != nil {
		t.Fatal(err)
	}
	payload = []byte(strings.Replace(string(payload), "0123456789abcdef0123456789abcdef01234567", head, 1))
	if err := os.WriteFile(eventPath, payload, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	env := testEnvironment(eventPath, server.URL, "")
	if err := run(root, []string{"-config", configPath}, &stdout, &stderr, env); err != nil {
		t.Fatalf("run() error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"blocking":false`) || !strings.Contains(stdout.String(), `"autoMergeEligible":false`) {
		t.Fatalf("decision = %s", stdout.String())
	}
}

func TestRunLeavesUnclearedPlatformBlocked(t *testing.T) {
	root, eventPath, configPath := ownerGateFixture(t, "OtherUser")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case strings.HasSuffix(request.URL.Path, "/pulls/42/files"):
			_, _ = io.WriteString(writer, `[{"filename":"products/us3/product.go","status":"modified"}]`)
		case strings.HasSuffix(request.URL.Path, "/collaborators/otheruser/permission"):
			_, _ = io.WriteString(writer, `{"permission":"pull"}`)
		case strings.HasSuffix(request.URL.Path, "/pulls/42/reviews"):
			_, _ = io.WriteString(writer, `[]`)
		default:
			t.Fatalf("unexpected API request: %s", request.URL)
		}
	}))
	defer server.Close()

	var stdout, stderr strings.Builder
	if err := run(root, []string{"-config", configPath}, &stdout, &stderr, testEnvironment(eventPath, server.URL, "")); err != nil {
		t.Fatalf("run() error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"blocking":true`) {
		t.Fatalf("decision = %s", stdout.String())
	}
}

func ownerGateFixture(t *testing.T, author string) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "product-owners.json")
	config := `{"version":1,"core":{"github_users":["CoreOwner"]},"products":{"us3":{"github_users":["ProductOwner"],"paths":["products/us3/**"]}}}`
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	eventPath := filepath.Join(root, "event.json")
	event := `{"repository":{"full_name":"ucloud/terraform-provider-ucloud"},"sender":{"login":"` + author + `"},"pull_request":{"number":42,"changed_files":1,"head":{"sha":"0123456789abcdef0123456789abcdef01234567"},"merge_commit_sha":"","user":{"login":"` + author + `"}}}`
	if err := os.WriteFile(eventPath, []byte(event), 0600); err != nil {
		t.Fatal(err)
	}
	return root, eventPath, configPath
}

func testEnvironment(eventPath, apiURL, outputPath string) func(string) string {
	return func(name string) string {
		switch name {
		case "GITHUB_EVENT_PATH":
			return eventPath
		case "GITHUB_API_URL":
			return apiURL
		case "GITHUB_TOKEN":
			return "test-token"
		case "GITHUB_OUTPUT":
			return outputPath
		default:
			return ""
		}
	}
}
