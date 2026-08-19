package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
				// the resource id is "<sec_group_id>:<rule_id>", Read parses
				// sec_group_id back out of it so no ImportStateIdFunc is needed
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
