package productownership

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type pullRequestReview struct {
	ID       int64  `json:"id"`
	State    string `json:"state"`
	CommitID string `json:"commit_id"`
	User     struct {
		Login string `json:"login"`
	} `json:"user"`
}

// RequireCoreApproval uses only the trusted policy and reviews of the exact
// proposed commit. New commits and dismissed approvals require a fresh review.
func (client GitHubClient) RequireCoreApproval(ctx context.Context, event PullRequestEvent, core Owners) error {
	if strings.TrimSpace(client.BaseURL) == "" || strings.TrimSpace(client.Token) == "" {
		return fmt.Errorf("core approval requires a GitHub API URL and token")
	}
	repository := strings.Split(event.Repository, "/")
	if len(repository) != 2 || !validRepositoryName(repository[0]) || !validRepositoryName(repository[1]) || event.Number < 1 || !validCommitSHA(event.HeadSHA) {
		return fmt.Errorf("invalid pull request identity for core approval")
	}
	owners, err := normalizeUsers(core.GitHubUsers)
	if err != nil || len(owners) == 0 {
		return fmt.Errorf("core approval requires valid trusted core maintainers")
	}
	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	latest := make(map[string]pullRequestReview)
	for page := 1; page <= 100; page++ {
		endpoint := strings.TrimRight(client.BaseURL, "/") + "/repos/" +
			url.PathEscape(repository[0]) + "/" + url.PathEscape(repository[1]) +
			"/pulls/" + strconv.Itoa(event.Number) + "/reviews?per_page=100&page=" + strconv.Itoa(page)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("create core approval request: %w", err)
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("Authorization", "Bearer "+client.Token)
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		response, err := httpClient.Do(request)
		if err != nil {
			return fmt.Errorf("request core approval reviews: %w", err)
		}
		if response.StatusCode != http.StatusOK {
			_ = response.Body.Close()
			return fmt.Errorf("core approval reviews returned %s", response.Status)
		}
		var reviews []pullRequestReview
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, 16<<20)).Decode(&reviews)
		closeErr := response.Body.Close()
		if decodeErr != nil {
			return fmt.Errorf("decode core approval reviews: %w", decodeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close core approval reviews: %w", closeErr)
		}
		for _, review := range reviews {
			login := normalizeUser(review.User.Login)
			if !contains(owners, login) || login == normalizeUser(event.Author) {
				continue
			}
			// Comments and pending reviews do not replace a submitted decision.
			if review.State == "COMMENTED" || review.State == "PENDING" {
				continue
			}
			if review.ID <= 0 {
				return fmt.Errorf("core approval review has an invalid ID")
			}
			if previous, exists := latest[login]; !exists || review.ID > previous.ID {
				latest[login] = review
			}
		}
		if len(reviews) < 100 {
			for _, review := range latest {
				if review.State == "APPROVED" && review.CommitID == event.HeadSHA {
					return nil
				}
			}
			return fmt.Errorf("product ownership changes require core approval of commit %s", event.HeadSHA)
		}
	}
	return fmt.Errorf("core approval review pagination exceeded 100 pages")
}
