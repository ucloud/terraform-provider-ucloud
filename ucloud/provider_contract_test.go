package ucloud

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
)

func TestProviderContract(t *testing.T) {
	provider := Provider().(*schema.Provider)
	metadata, err := productcatalog.AllOwnershipMetadata()
	if err != nil {
		t.Fatalf("read product catalog: %v", err)
	}
	baselines, err := providercompat.Baselines(provider, metadata)
	if err != nil {
		t.Fatalf("build provider contract baselines: %v", err)
	}

	names := make([]string, 0, len(baselines))
	for name := range baselines {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join("..", providercompat.BaselinePath(name))
			expected, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("read %s contract baseline %s: %v", name, filename, err)
			}
			actual := baselines[name]
			if !bytes.Equal(actual, expected) {
				line, expectedLine, actualLine := firstContractDifference(expected, actual)
				t.Fatalf(
					"%s contract differs from baseline %s at line %d\nexpected: %s\nactual:   %s\ninspect the difference and regenerate with `make contract PRODUCT=%s` if it is intentional (run this check through `make compat` or `make test`, which build the provider without inlining)",
					name,
					filename,
					line,
					expectedLine,
					actualLine,
					name,
				)
			}
		})
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
