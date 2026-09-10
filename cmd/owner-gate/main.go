package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
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
		fmt.Fprintf(os.Stderr, "owner gate failed: %v\n", err)
		os.Exit(1)
	}
}

func run(root string, args []string, stdout, stderr io.Writer, getenv func(string) string) error {
	flags := flag.NewFlagSet("owner-gate", flag.ContinueOnError)
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
	changes, err := client.PullRequestChanges(ctx, event)
	if err != nil {
		return fmt.Errorf("read pull request changes: %w", err)
	}
	decision, err := policy.ClassifyForMerge(event.Author, changes, false)
	if err != nil {
		return err
	}
	if decision.Type == productownership.MergeDecisionPlatform {
		cleared, err := client.RequireAdminApproval(ctx, event)
		if err != nil {
			return fmt.Errorf("read platform clearance: %w", err)
		}
		decision, err = policy.ClassifyForMerge(event.Author, changes, cleared)
		if err != nil {
			return err
		}
	}

	if err := writeDecisionOutput(getenv("GITHUB_OUTPUT"), decision); err != nil {
		return err
	}
	encoded, err := json.Marshal(decision)
	if err != nil {
		return fmt.Errorf("encode owner gate decision: %w", err)
	}
	fmt.Fprintln(stdout, string(encoded))
	return nil
}

func writeDecisionOutput(filename string, decision productownership.MergeDecision) error {
	if filename == "" {
		return nil
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open GitHub output file: %w", err)
	}
	defer file.Close()
	reason := strings.NewReplacer("\r", " ", "\n", " ").Replace(decision.Reason)
	if _, err := fmt.Fprintf(file, "type=%s\nproduct=%s\nautoMergeEligible=%t\nblocking=%t\nreason<<OWNER_GATE_EOF\n%s\nOWNER_GATE_EOF\n", decision.Type, decision.Product, decision.AutoMergeEligible, decision.Blocking, reason); err != nil {
		return fmt.Errorf("write GitHub output: %w", err)
	}
	return nil
}

func githubAPIURL(getenv func(string) string) string {
	if value := getenv("GITHUB_API_URL"); value != "" {
		return value
	}
	return "https://api.github.com"
}
