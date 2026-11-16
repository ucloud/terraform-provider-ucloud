package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
)

func TestRunWritesOnlySelectedProduct(test *testing.T) {
	for _, name := range []string{"uk8s", "core"} {
		test.Run(name, func(test *testing.T) {
			root := test.TempDir()
			protected := []string{
				providercompat.BaselinePath("us3"),
				providercompat.BaselinePath("core"),
			}
			if name == "core" {
				protected[1] = providercompat.BaselinePath("uk8s")
			}
			for _, filename := range protected {
				if err := writeBaseline(root, filename, []byte("existing baseline")); err != nil {
					test.Fatal(err)
				}
			}
			var stdout, stderr bytes.Buffer
			if err := run(root, []string{"-product", name}, &stdout, &stderr); err != nil {
				test.Fatal(err)
			}
			if stdout.Len() != 0 {
				test.Fatalf("unexpected stdout: %s", &stdout)
			}
			content, err := os.ReadFile(filepath.Join(root, providercompat.BaselinePath(name)))
			if err != nil {
				test.Fatal(err)
			}
			var contract providercompat.Contract
			if err := json.Unmarshal(content, &contract); err != nil {
				test.Fatal(err)
			}
			if name == "core" {
				if len(contract.Provider) == 0 || contract.Configure == "" || len(contract.Resources) != 0 || len(contract.DataSources) != 0 {
					test.Fatalf("unexpected core baseline: %#v", contract)
				}
			} else {
				metadata, err := productcatalog.OwnershipMetadataFor(name)
				if err != nil {
					test.Fatal(err)
				}
				if len(contract.Resources) != len(metadata.ResourceTypes) || len(contract.Provider) != 0 || contract.Configure != "" {
					test.Fatalf("unexpected product baseline: %#v", contract)
				}
				for _, resourceName := range metadata.ResourceTypes {
					if _, exists := contract.Resources[resourceName]; !exists {
						test.Fatalf("resource %s is missing from product baseline", resourceName)
					}
				}
			}
			for _, filename := range protected {
				content, err := os.ReadFile(filepath.Join(root, filename))
				if err != nil || string(content) != "existing baseline" {
					test.Fatalf("unselected baseline %s changed: %q, %v", filename, content, err)
				}
			}
		})
	}
}

func TestRunWritesAllBaselines(test *testing.T) {
	root := test.TempDir()
	if err := run(root, []string{"-product", "all"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		test.Fatal(err)
	}
	names := append(productcatalog.Names(), "core")
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(root, providercompat.BaselinePath(name)))
		if err != nil {
			test.Fatal(err)
		}
		if !json.Valid(content) {
			test.Fatalf("%s baseline is not valid JSON", name)
		}
	}
}

func TestRunPreservesFullExportAndSupportsExplicitOutput(test *testing.T) {
	for _, scenario := range []struct {
		name string
		args []string
	}{
		{name: "full contract stdout"},
		{name: "full contract file", args: []string{"-output", "export.json"}},
		{name: "product contract file", args: []string{"-product", "uk8s", "-output", "export.json"}},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			root := test.TempDir()
			var stdout, stderr bytes.Buffer
			if err := run(root, scenario.args, &stdout, &stderr); err != nil {
				test.Fatal(err)
			}
			content := stdout.Bytes()
			if len(scenario.args) != 0 {
				var err error
				content, err = os.ReadFile(filepath.Join(root, "export.json"))
				if err != nil {
					test.Fatal(err)
				}
				if stdout.Len() != 0 {
					test.Fatalf("unexpected stdout: %s", &stdout)
				}
			}
			var contract providercompat.Contract
			if err := json.Unmarshal(content, &contract); err != nil {
				test.Fatal(err)
			}
			if len(contract.Resources) == 0 {
				test.Fatal("export contains no resources")
			}
			if scenario.name != "product contract file" && (len(contract.Provider) == 0 || len(contract.DataSources) == 0) {
				test.Fatal("full export is missing provider configuration or data sources")
			}
		})
	}
}

func TestRunRejectsInvalidArgumentsWithoutWriting(test *testing.T) {
	for _, scenario := range []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown product", args: []string{"-product", "unknown"}, want: "unknown Provider product"},
		{name: "path traversal", args: []string{"-product", "../uk8s"}, want: "unknown Provider product"},
		{name: "all with output", args: []string{"-product", "all", "-output", "export.json"}, want: "-output cannot"},
		{name: "positional argument", args: []string{"uk8s"}, want: "unexpected arguments"},
		{name: "unknown flag", args: []string{"-unknown"}, want: "flag provided but not defined"},
	} {
		test.Run(scenario.name, func(test *testing.T) {
			root := test.TempDir()
			var stdout, stderr bytes.Buffer
			err := run(root, scenario.args, &stdout, &stderr)
			if err == nil || !strings.Contains(err.Error(), scenario.want) {
				test.Fatalf("run() error = %v, want %q", err, scenario.want)
			}
			files, err := os.ReadDir(root)
			if err != nil {
				test.Fatal(err)
			}
			if len(files) != 0 || stdout.Len() != 0 {
				test.Fatal("invalid arguments produced output")
			}
		})
	}
}

func TestRunHelp(test *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(test.TempDir(), []string{"-h"}, &stdout, &stderr)
	if !errors.Is(err, flag.ErrHelp) || !strings.Contains(stderr.String(), "-product") {
		test.Fatalf("help error = %v, output = %s", err, &stderr)
	}
}
