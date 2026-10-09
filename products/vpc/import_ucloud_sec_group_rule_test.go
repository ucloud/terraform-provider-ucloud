package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccUCloudSecGroupRule_import(t *testing.T) {
	resourceName := "ucloud_sec_group_rule.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupRuleConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// A rule cannot be located from its own ID alone, because
				// DescribeSecGroup has no RuleId filter, so the group it belongs
				// to travels with it in the import ID.
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					item, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("not found: %s", resourceName)
					}
					secGroupID := item.Primary.Attributes["sec_group_id"]
					if secGroupID == "" || item.Primary.ID == "" {
						return "", fmt.Errorf("sec group rule %q is missing an id", resourceName)
					}
					return fmt.Sprintf("%s/%s", secGroupID, item.Primary.ID), nil
				},
			},
		},
	})
}
