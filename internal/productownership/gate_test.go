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

func TestGatePublishesFailureForUnauthorizedPullRequest(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var states []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[{"filename":"products/us3/product.go","status":"modified"}]`))
		case http.MethodPost:
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode status request: %v", err)
			}
			states = append(states, body["state"])
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"id":1}`))
		default:
			t.Fatalf("unexpected method %q", request.Method)
		}
	}))
	defer server.Close()

	gate := productownership.Gate{
		Policy: policy,
		GitHub: productownership.GitHubClient{
			BaseURL:    server.URL,
			Token:      "test-token",
			HTTPClient: server.Client(),
		},
		StatusContext: "product-ownership",
		TargetURL:     "https://github.com/ucloud/terraform-provider-ucloud/actions/runs/1",
	}
	_, err = gate.Run(context.Background(), productownership.PullRequestEvent{
		Repository:   "ucloud/terraform-provider-ucloud",
		Number:       42,
		Author:       "OtherUser",
		Sender:       "OtherUser",
		ChangedFiles: 1,
		HeadSHA:      "0123456789abcdef0123456789abcdef01234567",
		MergeSHA:     "89abcdef0123456789abcdef0123456789abcdef",
	})
	if err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("Run() error = %v, want authorization error", err)
	}
	if got := strings.Join(states, ","); got != "pending,pending,failure,failure" {
		t.Fatalf("commit status states = %q, want pending,pending,failure,failure", got)
	}
}

func TestGatePublishesSuccessOnHeadAndMergeCommits(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var statuses []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[{"filename":"products/us3/product.go","status":"modified"}]`))
		case http.MethodPost:
			var body map[string]string
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode status request: %v", err)
			}
			statuses = append(statuses, body["state"]+":"+request.URL.Path)
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"id":1}`))
		default:
			t.Fatalf("unexpected method %q", request.Method)
		}
	}))
	defer server.Close()

	gate := productownership.Gate{
		Policy: policy,
		GitHub: productownership.GitHubClient{
			BaseURL:    server.URL,
			Token:      "test-token",
			HTTPClient: server.Client(),
		},
		StatusContext: "product-ownership",
	}
	headSHA := "0123456789abcdef0123456789abcdef01234567"
	mergeSHA := "89abcdef0123456789abcdef0123456789abcdef"
	result, err := gate.Run(context.Background(), productownership.PullRequestEvent{
		Repository:   "ucloud/terraform-provider-ucloud",
		Number:       42,
		Author:       "US3Owner",
		Sender:       "US3Owner",
		ChangedFiles: 1,
		HeadSHA:      headSHA,
		MergeSHA:     mergeSHA,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Decision.Owner != "us3" || result.ChangedFiles != 1 {
		t.Fatalf("Run() result = %#v", result)
	}
	want := strings.Join([]string{
		"pending:/repos/ucloud/terraform-provider-ucloud/statuses/" + headSHA,
		"pending:/repos/ucloud/terraform-provider-ucloud/statuses/" + mergeSHA,
		"success:/repos/ucloud/terraform-provider-ucloud/statuses/" + headSHA,
		"success:/repos/ucloud/terraform-provider-ucloud/statuses/" + mergeSHA,
	}, ",")
	if got := strings.Join(statuses, ","); got != want {
		t.Fatalf("commit statuses = %q, want %q", got, want)
	}
}

func TestGateRejectsUnauthorizedEventSender(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[{"filename":"products/us3/product.go","status":"modified"}]`))
		case http.MethodPost:
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"id":1}`))
		default:
			t.Fatalf("unexpected method %q", request.Method)
		}
	}))
	defer server.Close()

	gate := productownership.Gate{
		Policy: policy,
		GitHub: productownership.GitHubClient{
			BaseURL:    server.URL,
			Token:      "test-token",
			HTTPClient: server.Client(),
		},
		StatusContext: "product-ownership",
	}
	_, err = gate.Run(context.Background(), productownership.PullRequestEvent{
		Repository:   "ucloud/terraform-provider-ucloud",
		Number:       42,
		Author:       "US3Owner",
		Sender:       "OtherUser",
		ChangedFiles: 1,
		HeadSHA:      "0123456789abcdef0123456789abcdef01234567",
	})
	if err == nil || !strings.Contains(err.Error(), "event sender") {
		t.Fatalf("Run() error = %v, want unauthorized event sender error", err)
	}
}
