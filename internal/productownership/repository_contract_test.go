package productownership_test

import (
	"path/filepath"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/providercompat"
)

func TestProductContractBaselinesAllowAutonomousMerge(test *testing.T) {
	policy := loadRepositoryPolicy(test, repositoryRoot(test))
	policy.Core.GitHubUsers = []string{"coverage-core"}
	for _, name := range productcatalog.Names() {
		test.Run(name, func(test *testing.T) {
			changes := []productownership.Change{
				{Path: filepath.ToSlash(filepath.Join("products", name, "product.go"))},
				{Path: filepath.ToSlash(providercompat.BaselinePath(name))},
			}
			for _, author := range policy.Products[name].GitHubUsers {
				if _, err := policy.Authorize(author, changes); err != nil {
					test.Fatalf("%s cannot update %s code and baseline: %v", author, name, err)
				}
				decision, err := policy.ClassifyForMerge(author, changes, false)
				if err != nil {
					test.Fatal(err)
				}
				if decision.Product != name || !decision.AutoMergeEligible || decision.Blocking {
					test.Fatalf("product code and baseline require platform clearance: %#v", decision)
				}

				coreChanges := append(append([]productownership.Change{}, changes...), productownership.Change{
					Path: filepath.ToSlash(providercompat.BaselinePath("core")),
				})
				if _, err := policy.Authorize(author, coreChanges); err == nil {
					test.Fatalf("%s can change core baseline without core ownership", author)
				}
				decision, err = policy.ClassifyForMerge(author, coreChanges, false)
				if err != nil {
					test.Fatal(err)
				}
				if decision.AutoMergeEligible || !decision.Blocking {
					test.Fatalf("core baseline change bypassed platform clearance: %#v", decision)
				}
			}
		})
	}
}
