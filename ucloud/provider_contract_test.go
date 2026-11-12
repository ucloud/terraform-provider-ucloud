package ucloud

import (
	"bytes"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
)

// providerContractBaseline identifies the reviewed provider interface contract.
//
// The contract records schema hook names that providercompat resolves through
// runtime.FuncForPC, and a closure literal is named after the function that
// encloses it. Compiling the provider with inlining enabled can therefore
// report that closure under the name of the inlined caller instead, which made
// the baseline depend on the Go release. `make compat`, `make test` and `make
// contract` compile the provider with inlining disabled so the recorded names
// are stable; a bare `go test ./ucloud` is only comparable on a toolchain with
// the same inlining behaviour.
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
			"provider contract differs from baseline %s at line %d\nexpected: %s\nactual:   %s\npackage-only migrations must have zero contract changes; inspect the difference and regenerate with `make contract` if it is intentional (run this check through `make compat` or `make test`, which build the provider without inlining)",
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
