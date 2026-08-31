package productownership_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

func TestLoadPullRequestEvent(t *testing.T) {
	event, err := productownership.LoadPullRequestEvent(strings.NewReader(`{
		"repository": {"full_name": "ucloud/terraform-provider-ucloud"},
		"sender": {"login": "US3Owner"},
		"pull_request": {
			"number": 42,
			"changed_files": 2,
			"head": {"sha": "0123456789abcdef0123456789abcdef01234567"},
			"merge_commit_sha": "89abcdef0123456789abcdef0123456789abcdef",
			"user": {"login": "Ali1213"}
		}
	}`))
	if err != nil {
		t.Fatalf("LoadPullRequestEvent() error = %v", err)
	}
	if event.Repository != "ucloud/terraform-provider-ucloud" {
		t.Errorf("Repository = %q", event.Repository)
	}
	if event.Number != 42 {
		t.Errorf("Number = %d", event.Number)
	}
	if event.Author != "Ali1213" {
		t.Errorf("Author = %q", event.Author)
	}
	if event.Sender != "US3Owner" {
		t.Errorf("Sender = %q", event.Sender)
	}
	if event.ChangedFiles != 2 {
		t.Errorf("ChangedFiles = %d", event.ChangedFiles)
	}
	if event.HeadSHA != "0123456789abcdef0123456789abcdef01234567" {
		t.Errorf("HeadSHA = %q", event.HeadSHA)
	}
	if event.MergeSHA != "89abcdef0123456789abcdef0123456789abcdef" {
		t.Errorf("MergeSHA = %q", event.MergeSHA)
	}
}

func TestGitHubClientLoadsEveryPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		page := request.URL.Query().Get("page")
		files := make([]map[string]string, 0, 100)
		switch page {
		case "1":
			for index := 0; index < 100; index++ {
				files = append(files, map[string]string{"filename": "products/us3/generated.go", "status": "modified"})
			}
		case "2":
			files = append(files, map[string]string{"filename": "products/us3/final.go", "status": "added"})
		default:
			t.Fatalf("unexpected page %q", page)
		}
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(files); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client := productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}
	changes, err := client.PullRequestChanges(context.Background(), productownership.PullRequestEvent{
		Repository:   "ucloud/terraform-provider-ucloud",
		Number:       42,
		ChangedFiles: 101,
	})
	if err != nil {
		t.Fatalf("PullRequestChanges() error = %v", err)
	}
	if len(changes) != 101 {
		t.Fatalf("len(changes) = %d, want 101", len(changes))
	}
}

func TestGitHubClientRejectsIncompleteFileList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if request.URL.Query().Get("page") == "1" {
			_, _ = writer.Write([]byte(`[{"filename":"products/us3/product.go","status":"modified"}]`))
			return
		}
		_, _ = writer.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}
	_, err := client.PullRequestChanges(context.Background(), productownership.PullRequestEvent{
		Repository:   "ucloud/terraform-provider-ucloud",
		Number:       42,
		ChangedFiles: 2,
	})
	if err == nil || !strings.Contains(err.Error(), "1 of 2") {
		t.Fatalf("PullRequestChanges() error = %v, want incomplete file list error", err)
	}
}

func TestGitHubClientSetsCommitStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("request method = %q", request.Method)
		}
		if request.URL.Path != "/repos/ucloud/terraform-provider-ucloud/statuses/0123456789abcdef0123456789abcdef01234567" {
			t.Errorf("request path = %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		var body map[string]string
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body["state"] != "success" || body["context"] != "product-ownership" {
			t.Errorf("request body = %#v", body)
		}
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}
	err := client.SetCommitStatus(context.Background(), productownership.CommitStatus{
		Repository:  "ucloud/terraform-provider-ucloud",
		SHA:         "0123456789abcdef0123456789abcdef01234567",
		State:       "success",
		Context:     "product-ownership",
		Description: "product paths are authorized",
		TargetURL:   "https://github.com/ucloud/terraform-provider-ucloud/actions/runs/1",
	})
	if err != nil {
		t.Fatalf("SetCommitStatus() error = %v", err)
	}
}

func TestGitHubClientLoadsPullRequestChanges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/ucloud/terraform-provider-ucloud/pulls/42/files" {
			t.Errorf("request path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("per_page") != "100" || request.URL.Query().Get("page") != "1" {
			t.Errorf("request query = %q", request.URL.RawQuery)
		}
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[
			{"filename":"products/us3/product.go","status":"modified"},
			{"filename":"products/us3/new.go","previous_filename":"products/us3/old.go","status":"renamed"}
		]`))
	}))
	defer server.Close()

	client := productownership.GitHubClient{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: server.Client(),
	}
	changes, err := client.PullRequestChanges(context.Background(), productownership.PullRequestEvent{
		Repository:   "ucloud/terraform-provider-ucloud",
		Number:       42,
		Author:       "Ali1213",
		ChangedFiles: 2,
	})
	if err != nil {
		t.Fatalf("PullRequestChanges() error = %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("len(changes) = %d, want 2", len(changes))
	}
	if changes[0].Path != "products/us3/product.go" {
		t.Errorf("changes[0].Path = %q", changes[0].Path)
	}
	if changes[1].PreviousPath != "products/us3/old.go" {
		t.Errorf("changes[1].PreviousPath = %q", changes[1].PreviousPath)
	}
}
