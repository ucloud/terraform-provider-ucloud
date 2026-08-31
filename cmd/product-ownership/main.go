package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "product ownership check failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", ".github/product-owners.json", "trusted product ownership policy")
	eventPath := flag.String("event", os.Getenv("GITHUB_EVENT_PATH"), "GitHub pull request event payload")
	apiURL := flag.String("api-url", githubAPIURL(), "GitHub REST API base URL")
	flag.Parse()

	if *eventPath == "" {
		return fmt.Errorf("GitHub event path is empty; set GITHUB_EVENT_PATH or -event")
	}
	configFile, err := os.Open(*configPath)
	if err != nil {
		return fmt.Errorf("open product ownership policy %q: %w", *configPath, err)
	}
	policy, loadErr := productownership.Load(configFile)
	closeErr := configFile.Close()
	if loadErr != nil {
		return loadErr
	}
	if closeErr != nil {
		return fmt.Errorf("close product ownership policy %q: %w", *configPath, closeErr)
	}

	eventFile, err := os.Open(*eventPath)
	if err != nil {
		return fmt.Errorf("open GitHub event %q: %w", *eventPath, err)
	}
	event, loadErr := productownership.LoadPullRequestEvent(eventFile)
	closeErr = eventFile.Close()
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
		Token:   os.Getenv("GITHUB_TOKEN"),
	}
	gate := productownership.Gate{
		Policy:        policy,
		GitHub:        client,
		StatusContext: "product-ownership",
		TargetURL:     githubRunURL(),
	}
	result, err := gate.Run(ctx, event)
	if err != nil {
		return err
	}

	fmt.Printf(
		"authorized GitHub user %q as %q owner for %d changed files\n",
		event.Author,
		result.Decision.Owner,
		result.ChangedFiles,
	)
	return nil
}

func githubAPIURL() string {
	if value := os.Getenv("GITHUB_API_URL"); value != "" {
		return value
	}
	return "https://api.github.com"
}

func githubRunURL() string {
	server := os.Getenv("GITHUB_SERVER_URL")
	repository := os.Getenv("GITHUB_REPOSITORY")
	runID := os.Getenv("GITHUB_RUN_ID")
	if server == "" || repository == "" || runID == "" {
		return ""
	}
	return server + "/" + repository + "/actions/runs/" + runID
}
