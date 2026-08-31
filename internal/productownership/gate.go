package productownership

import (
	"context"
	"fmt"
)

// Gate connects the trusted policy, GitHub pull request data, and the required
// status written to the pull request head and test merge commits.
type Gate struct {
	Policy        *Policy
	GitHub        GitHubClient
	StatusContext string
	TargetURL     string
}

// GateResult describes one successfully authorized pull request.
type GateResult struct {
	Decision     Decision
	ChangedFiles int
}

// Run fails closed and publishes pending followed by success or failure.
func (gate Gate) Run(ctx context.Context, event PullRequestEvent) (GateResult, error) {
	if gate.Policy == nil {
		return GateResult{}, fmt.Errorf("product ownership policy is nil")
	}
	if gate.StatusContext == "" {
		return GateResult{}, fmt.Errorf("product ownership status context is empty")
	}
	status := CommitStatus{
		Repository:  event.Repository,
		State:       "pending",
		Context:     gate.StatusContext,
		Description: "checking pull request product ownership",
		TargetURL:   gate.TargetURL,
	}
	statusSHAs := pullRequestStatusSHAs(event)
	if err := gate.setCommitStatuses(ctx, status, statusSHAs); err != nil {
		return GateResult{}, fmt.Errorf("publish pending product ownership status: %w", err)
	}

	changes, checkErr := gate.GitHub.PullRequestChanges(ctx, event)
	var decision Decision
	if checkErr == nil {
		decision, checkErr = gate.Policy.Authorize(event.Author, changes)
	}
	if checkErr == nil && normalizeUser(event.Sender) != normalizeUser(event.Author) {
		if _, senderErr := gate.Policy.Authorize(event.Sender, changes); senderErr != nil {
			checkErr = fmt.Errorf("GitHub event sender %q is not authorized: %w", event.Sender, senderErr)
		}
	}
	if checkErr != nil {
		status.State = "failure"
		status.Description = "product ownership check failed"
		if statusErr := gate.setCommitStatuses(ctx, status, statusSHAs); statusErr != nil {
			return GateResult{}, fmt.Errorf("%v; publish failed product ownership status: %w", checkErr, statusErr)
		}
		return GateResult{}, checkErr
	}

	status.State = "success"
	status.Description = "product paths are authorized"
	if err := gate.setCommitStatuses(ctx, status, statusSHAs); err != nil {
		return GateResult{}, fmt.Errorf("publish successful product ownership status: %w", err)
	}
	return GateResult{Decision: decision, ChangedFiles: len(changes)}, nil
}

func (gate Gate) setCommitStatuses(ctx context.Context, status CommitStatus, shas []string) error {
	for _, sha := range shas {
		status.SHA = sha
		if err := gate.GitHub.SetCommitStatus(ctx, status); err != nil {
			return fmt.Errorf("commit %s: %w", sha, err)
		}
	}
	return nil
}

func pullRequestStatusSHAs(event PullRequestEvent) []string {
	shas := []string{event.HeadSHA}
	if event.MergeSHA != "" && event.MergeSHA != event.HeadSHA {
		shas = append(shas, event.MergeSHA)
	}
	return shas
}
