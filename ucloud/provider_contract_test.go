package ucloud

import (
	"bytes"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
)

const providerContractBaseline = "284a8f2e6a4dfc2a42be986791e62ac041c767f5"

func TestProviderContract(t *testing.T) {
	provider := Provider().(*schema.Provider)
	actual, err := providercompat.Marshal(provider)
	if err != nil {
		t.Fatalf("build provider contract: %v", err)
	}
	expected, err := os.ReadFile("testdata/provider_contract.json")
	if err != nil {
		t.Fatalf("read provider contract baseline %s: %v", providerContractBaseline, err)
	}

	if !bytes.Equal(actual, expected) {
		line, expectedLine, actualLine := firstContractDifference(expected, actual)
		t.Fatalf(
			"provider contract differs from baseline %s at line %d\nexpected: %s\nactual:   %s\npackage-only migrations must have zero contract changes; inspect and intentionally regenerate with `go run ./cmd/provider-contract -output ucloud/testdata/provider_contract.json`",
			providerContractBaseline,
			line,
			expectedLine,
			actualLine,
		)
	}
}

func firstContractDifference(expected, actual []byte) (int, string, string) {
	expectedLines := bytes.Split(expected, []byte{byte(10)})
	actualLines := bytes.Split(actual, []byte{byte(10)})
	lineCount := len(expectedLines)
	if len(actualLines) < lineCount {
		lineCount = len(actualLines)
	}
	for index := 0; index < lineCount; index++ {
		if !bytes.Equal(expectedLines[index], actualLines[index]) {
			return index + 1, string(expectedLines[index]), string(actualLines[index])
		}
	}
	if len(expectedLines) > lineCount {
		return lineCount + 1, string(expectedLines[lineCount]), "<missing>"
	}
	if len(actualLines) > lineCount {
		return lineCount + 1, "<missing>", string(actualLines[lineCount])
	}
	return 0, "", ""
}
