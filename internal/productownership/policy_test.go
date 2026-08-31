package productownership_test

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/productownership"
)

func TestProductOwnerCanChangeOwnedPaths(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["Ali1213"],
				"paths": [
					"products/us3/**",
					"examples/us3/**",
					"website/docs/d/us3_*.html.markdown"
				]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	decision, err := policy.Authorize("ali1213", []productownership.Change{
		{Path: "products/us3/resource_bucket.go"},
		{Path: "examples/us3/main.tf"},
		{Path: "website/docs/d/us3_bucket.html.markdown"},
	})
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if decision.Owner != "us3" {
		t.Fatalf("Authorize() owner = %q, want %q", decision.Owner, "us3")
	}
}

func TestLoadRejectsDuplicateGitHubUsers(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["Ali1213", "ali1213"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("Load() error = %v, want duplicate GitHub user error", err)
	}
}

func TestCoreMaintainerCannotBypassPathValidation(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["Ali1213"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = policy.Authorize("CoreMaintainer", []productownership.Change{{
		Path: "products/us3/../ucloud/provider.go",
	}})
	if err == nil || !strings.Contains(err.Error(), "repository-relative") {
		t.Fatalf("Authorize() error = %v, want repository-relative path error", err)
	}
}

func TestAuthorizeRejectsOwnershipBoundaryViolations(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner", "SharedOwner"],
				"paths": ["products/us3/**", "examples/us3/**"]
			},
			"uhost": {
				"github_users": ["UHostOwner", "SharedOwner"],
				"paths": ["products/uhost/**"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	tests := map[string]struct {
		author  string
		changes []productownership.Change
		want    string
	}{
		"different product owner": {
			author:  "UHostOwner",
			changes: []productownership.Change{{Path: "products/us3/product.go"}},
			want:    "not allowed",
		},
		"core path": {
			author:  "US3Owner",
			changes: []productownership.Change{{Path: ".github/product-owners.json"}},
			want:    "core maintainer",
		},
		"multiple products": {
			author: "SharedOwner",
			changes: []productownership.Change{
				{Path: "products/us3/product.go"},
				{Path: "products/uhost/product.go"},
			},
			want: "multiple products",
		},
		"rename across products": {
			author: "SharedOwner",
			changes: []productownership.Change{{
				Path:         "products/uhost/moved.go",
				PreviousPath: "products/us3/original.go",
			}},
			want: "multiple products",
		},
		"no changed files": {
			author: "US3Owner",
			want:   "no changed files",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := policy.Authorize(tc.author, tc.changes)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Authorize() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestCoreMaintainerCanChangeAnyValidRepositoryPath(t *testing.T) {
	policy, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	decision, err := policy.Authorize("coremaintainer", []productownership.Change{
		{Path: ".github/product-owners.json"},
		{Path: "ucloud/provider.go"},
	})
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if decision.Owner != "core" {
		t.Fatalf("Authorize() owner = %q, want core", decision.Owner)
	}
}

func TestLoadRequiresCoreMaintainer(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": []},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "core") {
		t.Fatalf("Load() error = %v, want missing core maintainer error", err)
	}
}

func TestLoadRejectsEmailAsGitHubUser(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["owner@example.com"],
				"paths": ["products/us3/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "GitHub user") {
		t.Fatalf("Load() error = %v, want invalid GitHub user error", err)
	}
}

func TestLoadRejectsCorePathAssignedToProduct(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": [".github/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "core-owned") {
		t.Fatalf("Load() error = %v, want core-owned path error", err)
	}
}

func TestLoadRejectsInvalidProductName(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"US3": {
				"github_users": ["US3Owner"],
				"paths": ["products/US3/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "product name") {
		t.Fatalf("Load() error = %v, want invalid product name error", err)
	}
}

func TestLoadRejectsPathAssignedToMultipleProducts(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["products/us3/**", "website/docs/shared/**"]
			},
			"uhost": {
				"github_users": ["UHostOwner"],
				"paths": ["products/uhost/**", "website/docs/shared/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "multiple products") {
		t.Fatalf("Load() error = %v, want multiple products path error", err)
	}
}

func TestLoadRequiresPrimaryProductPath(t *testing.T) {
	_, err := productownership.Load(strings.NewReader(`{
		"version": 1,
		"core": {"github_users": ["CoreMaintainer"]},
		"products": {
			"us3": {
				"github_users": ["US3Owner"],
				"paths": ["examples/us3/**"]
			}
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "products/us3/**") {
		t.Fatalf("Load() error = %v, want missing primary product path error", err)
	}
}
