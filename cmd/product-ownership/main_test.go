package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownershipsync"
)

func TestRunAuthorizesGeneratedProductOnboardingPolicy(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".github"), 0755); err != nil {
		t.Fatalf("create policy directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "products", "ulb"), 0755); err != nil {
		t.Fatalf("create product directory: %v", err)
	}
	baseline := []byte(`{
  "version": 1,
  "core": {
    "github_users": ["CoreOwner"]
  },
  "products": {
    "ulb": {
      "github_users": ["OldOwner"],
      "paths": ["products/ulb/**"]
    }
  }
}
`)
	policyPath := filepath.Join(root, ".github", "product-owners.json")
	if err := os.WriteFile(policyPath, baseline, 0644); err != nil {
		t.Fatalf("write baseline policy: %v", err)
	}
	generated, err := productownershipsync.GenerateFromPolicy(root, baseline, "ulb", []string{"NewOwner"})
	if err != nil {
		t.Fatalf("GenerateFromPolicy() error = %v", err)
	}

	var states []string
	contentRequested := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/pulls/42/files"):
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[{"filename":".github/product-owners.json","status":"modified"}]`))
		case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/contents/.github/product-owners.json"):
			contentRequested = true
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"type":     "file",
				"encoding": "base64",
				"content":  base64.StdEncoding.EncodeToString(generated.PolicyContents),
				"size":     len(generated.PolicyContents),
			})
		case request.Method == http.MethodPost && strings.Contains(request.URL.Path, "/statuses/"):
			var payload map[string]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode status payload: %v", err)
			}
			states = append(states, payload["state"])
			writer.WriteHeader(http.StatusCreated)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
		}
	}))
	defer server.Close()

	eventPath := filepath.Join(root, "event.json")
	event := `{
  "repository": {"full_name": "ucloud/terraform-provider-ucloud"},
  "sender": {"login": "NewOwner"},
  "pull_request": {
    "number": 42,
    "changed_files": 1,
    "head": {"sha": "0123456789abcdef0123456789abcdef01234567"},
    "user": {"login": "NewOwner"}
  }
}
`
	if err := os.WriteFile(eventPath, []byte(event), 0644); err != nil {
		t.Fatalf("write pull request event: %v", err)
	}
	environment := map[string]string{
		"GITHUB_API_URL":    server.URL,
		"GITHUB_TOKEN":      "test-token",
		"GITHUB_SERVER_URL": "https://github.example",
		"GITHUB_REPOSITORY": "ucloud/terraform-provider-ucloud",
		"GITHUB_RUN_ID":     "123",
		"GITHUB_EVENT_PATH": eventPath,
	}
	getenv := func(name string) string { return environment[name] }
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = run(
		root,
		[]string{"-config", policyPath, "-event", eventPath, "-api-url", server.URL},
		&stdout,
		&stderr,
		getenv,
	)
	if err != nil {
		t.Fatalf("run() error = %v, stderr = %s", err, stderr.String())
	}
	if !contentRequested {
		t.Fatal("run did not load the proposed policy file")
	}
	if got := strings.Join(states, ","); got != "pending,success" {
		t.Fatalf("status states = %q, want pending,success", got)
	}
	if !strings.Contains(stdout.String(), `authorized GitHub user "NewOwner" as "ulb" owner`) {
		t.Fatalf("run output = %q", stdout.String())
	}
}
