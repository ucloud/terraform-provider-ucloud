package productownership

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	maxPullRequestFiles       = 3000
	maxPullRequestFileContent = 1024 * 1024
)

// PullRequestEvent is the trusted identity and scope supplied by GitHub Actions.
type PullRequestEvent struct {
	Repository   string
	Number       int
	Author       string
	Sender       string
	ChangedFiles int
	HeadSHA      string
	MergeSHA     string
}

// GitHubClient reads pull request metadata without checking out or executing
// code from the pull request branch.
type GitHubClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// CommitStatus is a required status written to one pull request commit.
type CommitStatus struct {
	Repository  string
	SHA         string
	State       string
	Context     string
	Description string
	TargetURL   string
}

// LoadPullRequestEvent extracts only the pull request fields used by the gate.
func LoadPullRequestEvent(reader io.Reader) (PullRequestEvent, error) {
	var payload struct {
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Sender struct {
			Login string `json:"login"`
		} `json:"sender"`
		PullRequest struct {
			Number         int    `json:"number"`
			ChangedFiles   int    `json:"changed_files"`
			MergeCommitSHA string `json:"merge_commit_sha"`
			Head           struct {
				SHA string `json:"sha"`
			} `json:"head"`
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"pull_request"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		return PullRequestEvent{}, fmt.Errorf("decode GitHub pull request event: %w", err)
	}

	event := PullRequestEvent{
		Repository:   payload.Repository.FullName,
		Number:       payload.PullRequest.Number,
		Author:       payload.PullRequest.User.Login,
		Sender:       payload.Sender.Login,
		ChangedFiles: payload.PullRequest.ChangedFiles,
		HeadSHA:      payload.PullRequest.Head.SHA,
		MergeSHA:     payload.PullRequest.MergeCommitSHA,
	}
	parts := strings.Split(event.Repository, "/")
	if len(parts) != 2 || !validRepositoryName(parts[0]) || !validRepositoryName(parts[1]) {
		return PullRequestEvent{}, fmt.Errorf("GitHub repository %q must be owner/name", event.Repository)
	}
	if event.Number < 1 {
		return PullRequestEvent{}, fmt.Errorf("GitHub pull request number must be positive")
	}
	if !validGitHubUser(normalizeUser(event.Author)) {
		return PullRequestEvent{}, fmt.Errorf("GitHub pull request author %q is invalid", event.Author)
	}
	if !validGitHubUser(normalizeUser(event.Sender)) {
		return PullRequestEvent{}, fmt.Errorf("GitHub pull request event sender %q is invalid", event.Sender)
	}
	if event.ChangedFiles < 1 {
		return PullRequestEvent{}, fmt.Errorf("GitHub pull request changed_files must be positive")
	}
	if !validCommitSHA(event.HeadSHA) {
		return PullRequestEvent{}, fmt.Errorf("GitHub pull request head SHA %q is invalid", event.HeadSHA)
	}
	if event.MergeSHA != "" && !validCommitSHA(event.MergeSHA) {
		return PullRequestEvent{}, fmt.Errorf("GitHub pull request merge SHA %q is invalid", event.MergeSHA)
	}
	return event, nil
}

// PullRequestChanges loads every changed path and verifies GitHub returned the
// same count declared by the pull request event.
func (client GitHubClient) PullRequestChanges(ctx context.Context, event PullRequestEvent) ([]Change, error) {
	if strings.TrimSpace(client.BaseURL) == "" {
		return nil, fmt.Errorf("GitHub API base URL is empty")
	}
	if strings.TrimSpace(client.Token) == "" {
		return nil, fmt.Errorf("GitHub API token is empty")
	}
	if event.ChangedFiles < 1 {
		return nil, fmt.Errorf("GitHub pull request changed_files must be positive")
	}
	if event.ChangedFiles > maxPullRequestFiles {
		return nil, fmt.Errorf("GitHub pull request has %d changed files; maximum supported is %d", event.ChangedFiles, maxPullRequestFiles)
	}
	repository := strings.Split(event.Repository, "/")
	if len(repository) != 2 || !validRepositoryName(repository[0]) || !validRepositoryName(repository[1]) {
		return nil, fmt.Errorf("GitHub repository %q must be owner/name", event.Repository)
	}
	if event.Number < 1 {
		return nil, fmt.Errorf("GitHub pull request number must be positive")
	}

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	changes := make([]Change, 0, event.ChangedFiles)
	for pageNumber := 1; len(changes) < event.ChangedFiles; pageNumber++ {
		endpoint := strings.TrimRight(client.BaseURL, "/") +
			"/repos/" + url.PathEscape(repository[0]) +
			"/" + url.PathEscape(repository[1]) +
			"/pulls/" + strconv.Itoa(event.Number) +
			"/files?per_page=100&page=" + strconv.Itoa(pageNumber)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("create GitHub pull request files request: %w", err)
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("Authorization", "Bearer "+client.Token)
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		response, err := httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("request GitHub pull request files page %d: %w", pageNumber, err)
		}
		if response.StatusCode != http.StatusOK {
			message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
			_ = response.Body.Close()
			return nil, fmt.Errorf("GitHub pull request files page %d returned %s: %s", pageNumber, response.Status, strings.TrimSpace(string(message)))
		}
		var files []struct {
			Filename         string `json:"filename"`
			PreviousFilename string `json:"previous_filename"`
			Status           string `json:"status"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&files)
		closeErr := response.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("decode GitHub pull request files page %d: %w", pageNumber, decodeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close GitHub pull request files page %d: %w", pageNumber, closeErr)
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("GitHub returned %d of %d changed files", len(changes), event.ChangedFiles)
		}
		for _, file := range files {
			if file.Filename == "" {
				return nil, fmt.Errorf("GitHub returned a changed file without a filename")
			}
			if file.Status == "renamed" && file.PreviousFilename == "" {
				return nil, fmt.Errorf("GitHub returned renamed path %q without previous_filename", file.Filename)
			}
			changes = append(changes, Change{Path: file.Filename, PreviousPath: file.PreviousFilename})
		}
		if len(changes) > event.ChangedFiles {
			return nil, fmt.Errorf("GitHub returned %d files, more than event changed_files %d", len(changes), event.ChangedFiles)
		}
	}
	if len(changes) != event.ChangedFiles {
		return nil, fmt.Errorf("GitHub returned %d of %d changed files", len(changes), event.ChangedFiles)
	}
	return changes, nil
}

// PullRequestFileContent reads one file at the pull request head without
// checking out or executing content from the pull request branch.
func (client GitHubClient) PullRequestFileContent(ctx context.Context, event PullRequestEvent, filename string) ([]byte, error) {
	if strings.TrimSpace(client.BaseURL) == "" {
		return nil, fmt.Errorf("GitHub API base URL is empty")
	}
	if strings.TrimSpace(client.Token) == "" {
		return nil, fmt.Errorf("GitHub API token is empty")
	}
	repository := strings.Split(event.Repository, "/")
	if len(repository) != 2 || !validRepositoryName(repository[0]) || !validRepositoryName(repository[1]) {
		return nil, fmt.Errorf("GitHub repository %q must be owner/name", event.Repository)
	}
	if !validCommitSHA(event.HeadSHA) {
		return nil, fmt.Errorf("GitHub pull request head SHA %q is invalid", event.HeadSHA)
	}
	if err := validateRepositoryPath(filename); err != nil {
		return nil, fmt.Errorf("GitHub pull request filename %q: %w", filename, err)
	}

	escapedParts := strings.Split(filename, "/")
	for index := range escapedParts {
		escapedParts[index] = url.PathEscape(escapedParts[index])
	}
	endpoint := strings.TrimRight(client.BaseURL, "/") +
		"/repos/" + url.PathEscape(repository[0]) +
		"/" + url.PathEscape(repository[1]) +
		"/contents/" + strings.Join(escapedParts, "/") +
		"?ref=" + url.QueryEscape(event.HeadSHA)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create GitHub pull request file request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+client.Token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request GitHub pull request file: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("GitHub pull request file returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	var payload struct {
		Type     string `json:"type"`
		Encoding string `json:"encoding"`
		Content  string `json:"content"`
		Size     int    `json:"size"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 2*maxPullRequestFileContent+1))
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode GitHub pull request file: %w", err)
	}
	if payload.Type != "file" || payload.Encoding != "base64" {
		return nil, fmt.Errorf("GitHub pull request path %q is not a base64-encoded file", filename)
	}
	if payload.Size < 0 || payload.Size > maxPullRequestFileContent {
		return nil, fmt.Errorf("GitHub pull request file %q is %d bytes; maximum supported is %d", filename, payload.Size, maxPullRequestFileContent)
	}
	encoded := strings.Map(func(char rune) rune {
		if char == '\r' || char == '\n' || char == ' ' || char == '\t' {
			return -1
		}
		return char
	}, payload.Content)
	contents, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode GitHub pull request file %q content: %w", filename, err)
	}
	if len(contents) != payload.Size {
		return nil, fmt.Errorf("GitHub pull request file %q decoded to %d bytes, want %d", filename, len(contents), payload.Size)
	}
	return contents, nil
}

// SetCommitStatus publishes the trusted gate result on the pull request head.
func (client GitHubClient) SetCommitStatus(ctx context.Context, status CommitStatus) error {
	if strings.TrimSpace(client.BaseURL) == "" {
		return fmt.Errorf("GitHub API base URL is empty")
	}
	if strings.TrimSpace(client.Token) == "" {
		return fmt.Errorf("GitHub API token is empty")
	}
	repository := strings.Split(status.Repository, "/")
	if len(repository) != 2 || !validRepositoryName(repository[0]) || !validRepositoryName(repository[1]) {
		return fmt.Errorf("GitHub repository %q must be owner/name", status.Repository)
	}
	if !validCommitSHA(status.SHA) {
		return fmt.Errorf("GitHub commit SHA %q is invalid", status.SHA)
	}
	switch status.State {
	case "error", "failure", "pending", "success":
	default:
		return fmt.Errorf("GitHub commit status state %q is invalid", status.State)
	}
	if status.Context == "" {
		return fmt.Errorf("GitHub commit status context is empty")
	}

	payload, err := json.Marshal(map[string]string{
		"state":       status.State,
		"context":     status.Context,
		"description": status.Description,
		"target_url":  status.TargetURL,
	})
	if err != nil {
		return fmt.Errorf("encode GitHub commit status: %w", err)
	}
	endpoint := strings.TrimRight(client.BaseURL, "/") +
		"/repos/" + url.PathEscape(repository[0]) +
		"/" + url.PathEscape(repository[1]) +
		"/statuses/" + url.PathEscape(status.SHA)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create GitHub commit status request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+client.Token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("write GitHub commit status: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("GitHub commit status returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	return nil
}

func validRepositoryName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	for _, char := range name {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '-' && char != '_' && char != '.' {
			return false
		}
	}
	return true
}

func validCommitSHA(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
