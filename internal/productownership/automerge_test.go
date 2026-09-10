package productownership_test

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

func TestClassifyForMerge(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreOwner"]},
		"products": {
			"ulb": {"github_users": ["ProductOwner"], "paths": ["products/ulb/**"]},
			"us3": {"github_users": ["OtherProductOwner"], "paths": ["products/us3/**"]}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := []struct {
		name          string
		author        string
		changes       []productownership.Change
		platformClear bool
		wantType      string
		wantProduct   string
		wantAuto      bool
		wantBlocking  bool
	}{
		{
			name:        "product owner",
			author:      "productowner",
			changes:     []productownership.Change{{Path: "products/ulb/resource.go"}},
			wantType:    productownership.MergeDecisionProduct,
			wantProduct: "ulb",
			wantAuto:    true,
		},
		{
			name:         "non owner is blocked",
			author:       "otheruser",
			changes:      []productownership.Change{{Path: "products/ulb/resource.go"}},
			wantType:     productownership.MergeDecisionPlatform,
			wantBlocking: true,
		},
		{
			name:          "cleared platform",
			author:        "otheruser",
			changes:       []productownership.Change{{Path: "products/ulb/resource.go"}},
			platformClear: true,
			wantType:      productownership.MergeDecisionPlatform,
		},
		{
			name:         "cross product",
			author:       "productowner",
			changes:      []productownership.Change{{Path: "products/ulb/resource.go"}, {Path: "products/us3/resource.go"}},
			wantType:     productownership.MergeDecisionPlatform,
			wantBlocking: true,
		},
		{
			name:         "unowned path",
			author:       "productowner",
			changes:      []productownership.Change{{Path: ".github/workflows/auto-merge.yml"}},
			wantType:     productownership.MergeDecisionPlatform,
			wantBlocking: true,
		},
		{
			name:         "rename crosses boundary",
			author:       "productowner",
			changes:      []productownership.Change{{Path: "products/ulb/new.go", PreviousPath: ".github/old.go"}},
			wantType:     productownership.MergeDecisionPlatform,
			wantBlocking: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision, err := policy.ClassifyForMerge(test.author, test.changes, test.platformClear)
			if err != nil {
				t.Fatalf("ClassifyForMerge() error = %v", err)
			}
			if decision.Type != test.wantType || decision.Product != test.wantProduct || decision.AutoMergeEligible != test.wantAuto || decision.Blocking != test.wantBlocking {
				t.Fatalf("decision = %#v", decision)
			}
		})
	}
}

func TestClassifyForMergeRejectsInvalidPath(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreOwner"]},
		"products": {"ulb": {"github_users": ["ProductOwner"], "paths": ["products/ulb/**"]}}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, err := policy.ClassifyForMerge("ProductOwner", []productownership.Change{{Path: "../outside"}}, false); err == nil {
		t.Fatal("ClassifyForMerge() accepted a path traversal")
	}
}
