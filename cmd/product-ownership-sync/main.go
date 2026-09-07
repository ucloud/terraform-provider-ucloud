package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownershipsync"
)

func main() {
	root, err := os.Getwd()
	if err == nil {
		err = run(root, os.Args[1:], os.Stdout, os.Stderr)
	}
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "product ownership sync failed: %v\n", err)
		os.Exit(1)
	}
}

func run(root string, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("product-ownership-sync", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var githubUsers stringListFlag
	productName := flags.String("product", "", "exact Provider product name")
	flags.Var(&githubUsers, "github-user", "GitHub login to own the product; may be repeated")
	write := flags.Bool("write", false, "atomically update .github/product-owners.json")
	check := flags.Bool("check", false, "fail when the generated policy differs from the file")
	flags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s -product <name> -github-user <login> [-github-user <login>] [-write|-check]\n", flags.Name())
		fmt.Fprintln(stderr, "Without -write or -check, the command only previews the generated change.")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *productName == "" {
		return fmt.Errorf("-product is required")
	}
	if len(githubUsers) == 0 {
		return fmt.Errorf("at least one -github-user is required")
	}
	if *write && *check {
		return fmt.Errorf("-write and -check cannot be used together")
	}

	result, err := productownershipsync.Generate(root, *productName, githubUsers)
	if err != nil {
		return err
	}
	printResult(stdout, result)
	switch {
	case *check && result.Changed:
		return fmt.Errorf("%s is not synchronized; run again with -write", productownershipsync.PolicyRelativePath)
	case *check:
		fmt.Fprintf(stdout, "check: %s is synchronized\n", productownershipsync.PolicyRelativePath)
	case *write && result.Changed:
		if err := productownershipsync.Write(root, result.Baseline, result.PolicyContents); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "write: updated %s\n", productownershipsync.PolicyRelativePath)
	case *write:
		fmt.Fprintf(stdout, "write: %s already synchronized\n", productownershipsync.PolicyRelativePath)
	case result.Changed:
		fmt.Fprintf(stdout, "dry-run: changes required; rerun with -write to apply them\n")
	default:
		fmt.Fprintf(stdout, "dry-run: %s already synchronized\n", productownershipsync.PolicyRelativePath)
	}
	return nil
}

func printResult(writer io.Writer, result productownershipsync.Result) {
	fmt.Fprintf(writer, "product: %s\n", result.ProductName)
	fmt.Fprintf(writer, "master data: %s (%s)\n", result.MasterData.EnSampleName, result.MasterData.Key)
	if result.UsersChanged {
		fmt.Fprintf(writer, "- github_users: %s\n", strings.Join(result.Previous.GitHubUsers, ", "))
		fmt.Fprintf(writer, "+ github_users: %s\n", strings.Join(result.Generated.GitHubUsers, ", "))
	} else {
		fmt.Fprintf(writer, "  github_users: %s\n", strings.Join(result.Generated.GitHubUsers, ", "))
	}
	for _, ownedPath := range result.RemovedPaths {
		fmt.Fprintf(writer, "- path: %s\n", ownedPath)
	}
	for _, ownedPath := range result.AddedPaths {
		fmt.Fprintf(writer, "+ path: %s\n", ownedPath)
	}
	if len(result.AddedPaths) == 0 && len(result.RemovedPaths) == 0 {
		fmt.Fprintf(writer, "  paths: unchanged (%d)\n", len(result.Generated.Paths))
	} else {
		fmt.Fprintf(
			writer,
			"  paths: %d generated (%d added, %d removed)\n",
			len(result.Generated.Paths),
			len(result.AddedPaths),
			len(result.RemovedPaths),
		)
	}
}

type stringListFlag []string

func (values *stringListFlag) String() string {
	return strings.Join(*values, ",")
}

func (values *stringListFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}
