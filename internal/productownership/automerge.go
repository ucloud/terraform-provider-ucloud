package productownership

import (
	"fmt"
	"sort"
	"strings"
)

const (
	MergeDecisionProduct  = "product"
	MergeDecisionPlatform = "platform"
)

// MergeDecision routes a pull request to product autonomy or platform review.
// Platform decisions are blocking until an administrator clears the pull request.
type MergeDecision struct {
	Type              string `json:"type"`
	Product           string `json:"product,omitempty"`
	AutoMergeEligible bool   `json:"autoMergeEligible"`
	Blocking          bool   `json:"blocking"`
	Reason            string `json:"reason"`
}

// ClassifyForMerge uses the trusted base policy to determine whether a pull
// request is eligible for native GitHub auto-merge.
func (policy *Policy) ClassifyForMerge(author string, changes []Change, platformCleared bool) (MergeDecision, error) {
	if policy == nil {
		return MergeDecision{}, fmt.Errorf("product ownership policy is nil")
	}
	author = normalizeUser(author)
	if author == "" || !validGitHubUser(author) {
		return MergeDecision{}, fmt.Errorf("pull request author %q is invalid", author)
	}
	if len(changes) == 0 {
		return MergeDecision{}, fmt.Errorf("pull request has no changed files")
	}

	products := make(map[string]bool)
	platformPaths := make(map[string]bool)
	for _, change := range changes {
		for _, changedPath := range []string{change.Path, change.PreviousPath} {
			if changedPath == "" {
				continue
			}
			if err := validateRepositoryPath(changedPath); err != nil {
				return MergeDecision{}, fmt.Errorf("changed path %q: %w", changedPath, err)
			}
			product, found, err := policy.ProductOwner(changedPath)
			if err != nil {
				return MergeDecision{}, err
			}
			if !found {
				platformPaths[changedPath] = true
				continue
			}
			products[product] = true
		}
	}

	if len(platformPaths) > 0 || len(products) != 1 {
		detail := "the pull request touches platform-owned paths"
		if len(products) > 1 {
			names := make([]string, 0, len(products))
			for product := range products {
				names = append(names, product)
			}
			sort.Strings(names)
			detail = fmt.Sprintf("the pull request touches multiple products: %s", strings.Join(names, ", "))
		}
		if len(platformPaths) > 0 {
			paths := make([]string, 0, len(platformPaths))
			for changedPath := range platformPaths {
				paths = append(paths, changedPath)
			}
			sort.Strings(paths)
			detail = fmt.Sprintf("the pull request touches platform-owned paths: %s", strings.Join(paths, ", "))
		}
		return platformMergeDecision(detail, platformCleared), nil
	}

	product := ""
	for name := range products {
		product = name
	}
	if !contains(policy.Products[product].GitHubUsers, author) {
		return platformMergeDecision(
			fmt.Sprintf("%s is not a configured owner of product %s", author, product),
			platformCleared,
		), nil
	}

	return MergeDecision{
		Type:              MergeDecisionProduct,
		Product:           product,
		AutoMergeEligible: true,
		Reason:            fmt.Sprintf("%s is a configured owner of product %s", author, product),
	}, nil
}

func platformMergeDecision(detail string, platformCleared bool) MergeDecision {
	decision := MergeDecision{
		Type:     MergeDecisionPlatform,
		Reason:   detail,
		Blocking: !platformCleared,
	}
	if platformCleared {
		decision.Reason = "platform review cleared: " + detail
	} else {
		decision.Reason = "platform review required: " + detail
	}
	return decision
}
