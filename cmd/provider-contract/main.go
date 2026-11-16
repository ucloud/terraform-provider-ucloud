package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
	"github.com/terraform-providers/terraform-provider-ucloud/ucloud"
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
		fmt.Fprintf(os.Stderr, "provider contract: %v\n", err)
		os.Exit(1)
	}
}

func run(root string, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("provider-contract", flag.ContinueOnError)
	flags.SetOutput(stderr)
	productName := flags.String("product", "", "write one product baseline, core, or all baselines; omit to export the full contract")
	output := flags.String("output", "", "output file; defaults to stdout for a full export or the selected product's baseline path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *productName == "all" && *output != "" {
		return fmt.Errorf("-output cannot be combined with -product all")
	}
	if *productName != "" && *productName != "core" && *productName != "all" {
		if _, err := productcatalog.OwnershipMetadataFor(*productName); err != nil {
			return err
		}
	}

	provider := ucloud.Provider().(*schema.Provider)
	if *productName == "" {
		contract, err := providercompat.Marshal(provider)
		if err != nil {
			return fmt.Errorf("build provider contract: %w", err)
		}
		if *output == "" {
			_, err = stdout.Write(contract)
			return err
		}
		return writeBaseline(root, *output, contract)
	}

	metadata, err := productcatalog.AllOwnershipMetadata()
	if err != nil {
		return err
	}
	baselines, err := providercompat.Baselines(provider, metadata)
	if err != nil {
		return err
	}
	if *productName != "all" {
		filename := *output
		if filename == "" {
			filename = providercompat.BaselinePath(*productName)
		}
		return writeBaseline(root, filename, baselines[*productName])
	}

	names := make([]string, 0, len(baselines))
	for name := range baselines {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := writeBaseline(root, providercompat.BaselinePath(name), baselines[name]); err != nil {
			return err
		}
	}
	return nil
}

func writeBaseline(root, filename string, contract []byte) error {
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(root, filename)
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("create baseline directory: %w", err)
	}
	if err := os.WriteFile(filename, contract, 0644); err != nil {
		return fmt.Errorf("write baseline %q: %w", filename, err)
	}
	return nil
}
