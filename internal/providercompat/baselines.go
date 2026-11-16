package providercompat

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
)

var baselineProductName = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func BaselinePath(productName string) string {
	if productName == "core" {
		return filepath.Join("ucloud", "testdata", "provider_contract.json")
	}
	return filepath.Join("products", productName, "testdata", "provider_contract.json")
}

func Baselines(provider *schema.Provider, products []productcatalog.OwnershipMetadata) (map[string][]byte, error) {
	contract, err := Build(provider)
	if err != nil {
		return nil, err
	}
	contracts := map[string]Contract{
		"core": {
			FormatVersion: contract.FormatVersion,
			Provider:      contract.Provider,
			Configure:     contract.Configure,
		},
	}
	for _, metadata := range products {
		if !baselineProductName.MatchString(metadata.Name) || metadata.Name == "core" || metadata.Name == "all" {
			return nil, fmt.Errorf("invalid baseline product name %q", metadata.Name)
		}
		if _, exists := contracts[metadata.Name]; exists {
			return nil, fmt.Errorf("duplicate baseline product %q", metadata.Name)
		}
		resources, err := takeContracts(contract.Resources, metadata.ResourceTypes)
		if err != nil {
			return nil, fmt.Errorf("product %q resources: %w", metadata.Name, err)
		}
		dataSources, err := takeContracts(contract.DataSources, metadata.DataSourceTypes)
		if err != nil {
			return nil, fmt.Errorf("product %q data sources: %w", metadata.Name, err)
		}
		contracts[metadata.Name] = Contract{
			FormatVersion: contract.FormatVersion,
			Resources:     resources,
			DataSources:   dataSources,
		}
	}
	if len(contract.Resources) > 0 {
		return nil, fmt.Errorf("provider resources have no product owner: %s", contractNames(contract.Resources))
	}
	if len(contract.DataSources) > 0 {
		return nil, fmt.Errorf("provider data sources have no product owner: %s", contractNames(contract.DataSources))
	}

	baselines := make(map[string][]byte, len(contracts))
	for name, partition := range contracts {
		encoded, err := marshalContract(partition)
		if err != nil {
			return nil, fmt.Errorf("product %q baseline: %w", name, err)
		}
		baselines[name] = encoded
	}
	return baselines, nil
}

func takeContracts(source map[string]ResourceContract, names []string) (map[string]ResourceContract, error) {
	result := make(map[string]ResourceContract, len(names))
	for _, name := range names {
		contract, exists := source[name]
		if !exists {
			return nil, fmt.Errorf("%q is not registered or is assigned more than once", name)
		}
		result[name] = contract
		delete(source, name)
	}
	return result, nil
}

func contractNames(contracts map[string]ResourceContract) string {
	names := make([]string, 0, len(contracts))
	for name := range contracts {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
