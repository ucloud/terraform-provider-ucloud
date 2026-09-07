package us3

import (
	"os"
	"testing"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

func TestUpgradeFixtureSyntax(t *testing.T) {
	parser := hclparse.NewParser()
	if _, diagnostics := parser.ParseHCLFile("test-fixtures/upgrade/main.tf"); diagnostics.HasErrors() {
		t.Errorf("parse Terraform fixture: %s", diagnostics.Error())
	}

	cliConfig, err := os.ReadFile("test-fixtures/upgrade/dev.tfrc")
	if err != nil {
		t.Fatalf("read Terraform CLI fixture: %v", err)
	}
	if _, err := hcl.Parse(string(cliConfig)); err != nil {
		t.Errorf("parse Terraform CLI fixture: %v", err)
	}
}
