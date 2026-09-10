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

// RequireAdminApproval returns true when the author is an administrator or an
// administrator other than the author approved the current head commit.
func (client GitHubClient) RequireAdminApproval(ctx context.Context, event PullRequestEvent) (bool, error) {
	if strings.TrimSpace(client.BaseURL) == "" || strings.TrimSpace(client.Token) == "" {
		return false, fmt.Errorf("administrator clearance requires a GitHub API URL and token")
	}
	repository := strings.Split(event.Repository, "/")
	if len(repository) != 2 || !validRepositoryName(repository[0]) || !validRepositoryName(repository[1]) || event.Number < 1 || !validCommitSHA(event.HeadSHA) {
		return false, fmt.Errorf("invalid pull request identity for administrator clearance")
	}

	authorPermission, err := client.collaboratorPermission(ctx, repository[0], repository[1], event.Author)
	if err != nil {
		return false, fmt.Errorf("read pull request author permission: %w", err)
	}
	if elevatedPermission(authorPermission) {
		return true, nil
	}

	reviews, err := client.latestPullRequestReviews(ctx, event, repository[0], repository[1])
	if err != nil {
		return false, err
	}
	for login, review := range reviews {
		if login == normalizeUser(event.Author) || review.State != "APPROVED" || review.CommitID != event.HeadSHA {
			continue
		}
		permission, err := client.collaboratorPermission(ctx, repository[0], repository[1], login)
		if err != nil {
			return false, fmt.Errorf("read reviewer %q permission: %w", login, err)
		}
		if elevatedPermission(permission) {
			return true, nil
		}
	}
	return false, nil
}

func (client GitHubClient) latestPullRequestReviews(ctx context.Context, event PullRequestEvent, owner, repository string) (map[string]pullRequestReview, error) {
	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	latest := make(map[string]pullRequestReview)
	for page := 1; page <= 100; page++ {
		endpoint := strings.TrimRight(client.BaseURL, "/") + "/repos/" +
			url.PathEscape(owner) + "/" + url.PathEscape(repository) +
			"/pulls/" + strconv.Itoa(event.Number) + "/reviews?per_page=100&page=" + strconv.Itoa(page)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("create administrator review request: %w", err)
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("Authorization", "Bearer "+client.Token)
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		response, err := httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("request administrator reviews: %w", err)
		}
		if response.StatusCode != http.StatusOK {
			message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
			_ = response.Body.Close()
			return nil, fmt.Errorf("administrator reviews returned %s: %s", response.Status, strings.TrimSpace(string(message)))
		}
		var reviews []pullRequestReview
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, 16<<20)).Decode(&reviews)
		closeErr := response.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("decode administrator reviews: %w", decodeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close administrator reviews: %w", closeErr)
		}
		for _, review := range reviews {
			login := normalizeUser(review.User.Login)
			if login == "" {
				return nil, fmt.Errorf("administrator review has an empty reviewer")
			}
			if review.State == "COMMENTED" || review.State == "PENDING" {
				continue
			}
			if review.ID <= 0 {
				return nil, fmt.Errorf("administrator review has an invalid ID")
			}
			if previous, exists := latest[login]; !exists || review.ID > previous.ID {
				latest[login] = review
			}
		}
		if len(reviews) < 100 {
			return latest, nil
		}
	}
	return nil, fmt.Errorf("administrator review pagination exceeded 100 pages")
}

func (client GitHubClient) collaboratorPermission(ctx context.Context, owner, repository, login string) (string, error) {
	login = normalizeUser(login)
	if !validGitHubUser(login) {
		return "", fmt.Errorf("GitHub user %q is invalid", login)
	}
	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	endpoint := strings.TrimRight(client.BaseURL, "/") + "/repos/" +
		url.PathEscape(owner) + "/" + url.PathEscape(repository) +
		"/collaborators/" + url.PathEscape(login) + "/permission"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create collaborator permission request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+client.Token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("request collaborator permission: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return "", fmt.Errorf("collaborator permission returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	var payload struct {
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode collaborator permission: %w", err)
	}
	return payload.Permission, nil
}

func elevatedPermission(permission string) bool {
	return permission == "admin" || permission == "maintain"
}
