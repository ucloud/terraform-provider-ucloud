package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDryRunWriteAndCheck(t *testing.T) {
	root := commandRepositoryFixture(t)
	policyPath := filepath.Join(root, ".github", "product-owners.json")
	before, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read policy before dry-run: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	args := []string{"-product", "ulb", "-github-user", "NewOwner"}
	if err := run(root, args, &stdout, &stderr); err != nil {
		t.Fatalf("run(dry-run) error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "dry-run: changes required") ||
		!strings.Contains(stdout.String(), "+ github_users: NewOwner") {
		t.Fatalf("dry-run output = %q", stdout.String())
	}
	afterDryRun, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read policy after dry-run: %v", err)
	}
	if !bytes.Equal(before, afterDryRun) {
		t.Fatal("dry-run changed product-owners.json")
	}

	stdout.Reset()
	stderr.Reset()
	if err := run(root, append(args, "-write"), &stdout, &stderr); err != nil {
		t.Fatalf("run(write) error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "write: updated .github/product-owners.json") {
		t.Fatalf("write output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if err := run(root, append(args, "-check"), &stdout, &stderr); err != nil {
		t.Fatalf("run(check) error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "check: .github/product-owners.json is synchronized") {
		t.Fatalf("check output = %q", stdout.String())
	}
}

func TestRunCheckFailsForDrift(t *testing.T) {
	root := commandRepositoryFixture(t)
	err := run(
		root,
		[]string{"-product", "ulb", "-github-user", "DifferentOwner", "-check"},
		&bytes.Buffer{},
		&bytes.Buffer{},
	)
	if err == nil || !strings.Contains(err.Error(), "is not synchronized") {
		t.Fatalf("run(check drift) error = %v, want synchronization error", err)
	}
}

func TestRunHelpAndFlagValidation(t *testing.T) {
	var output bytes.Buffer
	if err := run(t.TempDir(), []string{"-h"}, &bytes.Buffer{}, &output); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("run(-h) error = %v, want flag.ErrHelp", err)
	}
	if !strings.Contains(output.String(), "Usage: product-ownership-sync") || !strings.Contains(output.String(), "-write") {
		t.Fatalf("help output = %q", output.String())
	}
	if err := run(t.TempDir(), []string{"-product", "ulb"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "-github-user") {
		t.Fatalf("run(missing user) error = %v", err)
	}
	if err := run(
		t.TempDir(),
		[]string{"-product", "ulb", "-github-user", "Owner", "-write", "-check"},
		&bytes.Buffer{},
		&bytes.Buffer{},
	); err == nil || !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("run(conflicting modes) error = %v", err)
	}
}

func commandRepositoryFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{".github", "products/ulb"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(directory)), 0755); err != nil {
			t.Fatalf("create fixture directory %s: %v", directory, err)
		}
	}
	policy := `{
  "version": 1,
  "core": {
    "github_users": ["CoreOwner"]
  },
  "products": {
    "ulb": {
      "github_users": ["OldOwner"],
      "paths": ["products/ulb/**"]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(root, ".github", "product-owners.json"), []byte(policy), 0644); err != nil {
		t.Fatalf("write fixture policy: %v", err)
	}
	return root
}
