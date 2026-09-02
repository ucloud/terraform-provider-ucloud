package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownershipsync"
)

func main() {
	root, err := os.Getwd()
	if err == nil {
		err = run(root, os.Args[1:], os.Stdout, os.Stderr, os.Getenv)
	}
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "product ownership check failed: %v\n", err)
		os.Exit(1)
	}
}

func run(root string, args []string, stdout, stderr io.Writer, getenv func(string) string) error {
	flags := flag.NewFlagSet("product-ownership", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", ".github/product-owners.json", "trusted product ownership policy")
	eventPath := flags.String("event", getenv("GITHUB_EVENT_PATH"), "GitHub pull request event payload")
	apiURL := flags.String("api-url", githubAPIURL(getenv), "GitHub REST API base URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}

	if *eventPath == "" {
		return fmt.Errorf("GitHub event path is empty; set GITHUB_EVENT_PATH or -event")
	}
	configContents, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("read product ownership policy %q: %w", *configPath, err)
	}
	policy, err := productownership.Load(bytes.NewReader(configContents))
	if err != nil {
		return err
	}

	eventFile, err := os.Open(*eventPath)
	if err != nil {
		return fmt.Errorf("open GitHub event %q: %w", *eventPath, err)
	}
	event, loadErr := productownership.LoadPullRequestEvent(eventFile)
	closeErr := eventFile.Close()
	if loadErr != nil {
		return loadErr
	}
	if closeErr != nil {
		return fmt.Errorf("close GitHub event %q: %w", *eventPath, closeErr)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	client := productownership.GitHubClient{
		BaseURL: *apiURL,
		Token:   getenv("GITHUB_TOKEN"),
	}
	gate := productownership.Gate{
		Policy: policy,
		GitHub: client,
		OnboardingAuthorizer: func(ctx context.Context, event productownership.PullRequestEvent) (productownership.Decision, error) {
			proposal, err := client.PullRequestFileContent(ctx, event, productownershipsync.PolicyRelativePath)
			if err != nil {
				return productownership.Decision{}, err
			}
			owner, err := productownershipsync.ValidateProposal(
				root,
				configContents,
				proposal,
				event.Author,
				event.Sender,
			)
			if err != nil {
				return productownership.Decision{}, err
			}
			return productownership.Decision{Owner: owner}, nil
		},
		StatusContext: "product-ownership",
		TargetURL:     githubRunURL(getenv),
	}
	result, err := gate.Run(ctx, event)
	if err != nil {
		return err
	}

	fmt.Fprintf(
		stdout,
		"authorized GitHub user %q as %q owner for %d changed files\n",
		event.Author,
		result.Decision.Owner,
		result.ChangedFiles,
	)
	return nil
}

func githubAPIURL(getenv func(string) string) string {
	if value := getenv("GITHUB_API_URL"); value != "" {
		return value
	}
	return "https://api.github.com"
}

func githubRunURL(getenv func(string) string) string {
	server := getenv("GITHUB_SERVER_URL")
	repository := getenv("GITHUB_REPOSITORY")
	runID := getenv("GITHUB_RUN_ID")
	if server == "" || repository == "" || runID == "" {
		return ""
	}
	return server + "/" + repository + "/actions/runs/" + runID
}
