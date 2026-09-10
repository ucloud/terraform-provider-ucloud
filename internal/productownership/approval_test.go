package productownership_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

func TestCoreApprovalReadsAllPagesBeforeAccepting(t *testing.T) {
	head := strings.Repeat("a", 40)
	for _, test := range []struct {
		name  string
		state string
		allow bool
	}{
		{"approval after first page", "APPROVED", true},
		{"later dismissal", "DISMISSED", false},
		{"later rejection", "CHANGES_REQUESTED", false},
		{"comment preserves approval", "COMMENTED", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			pages := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				pages++
				if r.Header.Get("Authorization") != "Bearer test-token" || r.URL.Path != "/repos/ucloud/terraform-provider-ucloud/pulls/42/reviews" {
					t.Errorf("unexpected review request: %s", r.URL)
				}
				review := func(id int, login, state string) map[string]interface{} {
					return map[string]interface{}{"id": id, "user": map[string]string{"login": login}, "state": state, "commit_id": head}
				}
				var reviews []map[string]interface{}
				switch r.URL.Query().Get("page") {
				case "1":
					for i := 1; i <= 100; i++ {
						reviews = append(reviews, review(i, "OtherUser", "COMMENTED"))
					}
					if test.state != "APPROVED" {
						reviews[0] = review(1, "COREOWNER", "APPROVED")
					}
				case "2":
					reviews = []map[string]interface{}{review(101, "CoreOwner", test.state)}
				default:
					t.Errorf("unexpected page: %s", r.URL)
				}
				_ = json.NewEncoder(w).Encode(reviews)
			}))
			defer server.Close()
			client := productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}
			err := client.RequireCoreApproval(context.Background(), productownership.PullRequestEvent{
				Repository: "ucloud/terraform-provider-ucloud", Number: 42, Author: "NewOwner", HeadSHA: head,
			}, productownership.Owners{GitHubUsers: []string{"CoreOwner"}})
			if (err == nil) != test.allow {
				t.Fatalf("RequireCoreApproval() = %v, want allowed=%t", err, test.allow)
			}
			if pages != 2 {
				t.Fatalf("read %d pages, want 2", pages)
			}
		})
	}
}

func TestCoreApprovalFailsClosedOnAPIError(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusInternalServerError, http.StatusOK} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("invalid JSON"))
			}))
			defer server.Close()
			client := productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}
			err := client.RequireCoreApproval(context.Background(), productownership.PullRequestEvent{
				Repository: "ucloud/terraform-provider-ucloud", Number: 42, Author: "NewOwner", HeadSHA: strings.Repeat("a", 40),
			}, productownership.Owners{GitHubUsers: []string{"CoreOwner"}})
			if err == nil {
				t.Fatal("API failure must not authorize an ownership change")
			}
		})
	}
}

func TestRequireAdminApproval(t *testing.T) {
	head := strings.Repeat("a", 40)
	for _, test := range []struct {
		name        string
		author      string
		reviews     string
		permissions map[string]string
		wantCleared bool
		wantPages   int
	}{
		{
			name:        "admin author",
			author:      "AdminAuthor",
			permissions: map[string]string{"adminauthor": "admin"},
			wantCleared: true,
		},
		{
			name:        "current admin approval",
			author:      "Author",
			reviews:     `[{"id":1,"state":"APPROVED","commit_id":"` + head + `","user":{"login":"AdminReviewer"}}]`,
			permissions: map[string]string{"author": "pull", "adminreviewer": "maintain"},
			wantCleared: true,
			wantPages:   1,
		},
		{
			name:        "old approval does not clear",
			author:      "Author",
			reviews:     `[{"id":1,"state":"APPROVED","commit_id":"` + strings.Repeat("b", 40) + `","user":{"login":"AdminReviewer"}}]`,
			permissions: map[string]string{"author": "pull", "adminreviewer": "admin"},
			wantPages:   1,
		},
		{
			name:        "non admin approval does not clear",
			author:      "Author",
			reviews:     `[{"id":1,"state":"APPROVED","commit_id":"` + head + `","user":{"login":"ProductOwner"}}]`,
			permissions: map[string]string{"author": "pull", "productowner": "push"},
			wantPages:   1,
		},
		{
			name:        "later changes request invalidates approval",
			author:      "Author",
			reviews:     `[{"id":1,"state":"APPROVED","commit_id":"` + head + `","user":{"login":"AdminReviewer"}},{"id":2,"state":"CHANGES_REQUESTED","commit_id":"` + head + `","user":{"login":"AdminReviewer"}}]`,
			permissions: map[string]string{"author": "pull", "adminreviewer": "admin"},
			wantPages:   1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			pages := 0
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch {
				case strings.HasSuffix(request.URL.Path, "/reviews"):
					pages++
					_, _ = io.WriteString(writer, test.reviews)
				case strings.Contains(request.URL.Path, "/collaborators/") && strings.HasSuffix(request.URL.Path, "/permission"):
					parts := strings.Split(request.URL.Path, "/")
					login := parts[len(parts)-2]
					permission, ok := test.permissions[strings.ToLower(login)]
					if !ok {
						writer.WriteHeader(http.StatusNotFound)
						return
					}
					_, _ = io.WriteString(writer, `{"permission":"`+permission+`"}`)
				default:
					t.Fatalf("unexpected API request: %s", request.URL)
				}
			}))
			defer server.Close()

			cleared, err := (productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}).RequireAdminApproval(context.Background(), productownership.PullRequestEvent{
				Repository: "ucloud/terraform-provider-ucloud",
				Number:     42,
				Author:     test.author,
				HeadSHA:    head,
			})
			if err != nil {
				t.Fatalf("RequireAdminApproval() error = %v", err)
			}
			if cleared != test.wantCleared {
				t.Fatalf("RequireAdminApproval() = %t, want %t", cleared, test.wantCleared)
			}
			if pages != test.wantPages {
				t.Fatalf("review pages = %d, want %d", pages, test.wantPages)
			}
		})
	}
}

func TestRequireAdminApprovalFailsClosedOnReviewAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "/permission") {
			_, _ = io.WriteString(writer, `{"permission":"pull"}`)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(writer, "server failure")
	}))
	defer server.Close()

	_, err := (productownership.GitHubClient{BaseURL: server.URL, Token: "test-token", HTTPClient: server.Client()}).RequireAdminApproval(context.Background(), productownership.PullRequestEvent{
		Repository: "ucloud/terraform-provider-ucloud",
		Number:     42,
		Author:     "Author",
		HeadSHA:    strings.Repeat("a", 40),
	})
	if err == nil {
		t.Fatal("RequireAdminApproval() accepted a review API failure")
	}
}
