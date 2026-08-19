package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudSecGroupRule_basic(t *testing.T) {
	var ruleSet vpc.SecGroupRuleInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_sec_group_rule.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecGroupRuleDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupRuleConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &ruleSet),
					testAccCheckSecGroupRuleAttributes(&ruleSet),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "direction", "Ingress"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "protocol_type", "TCP"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "dst_port", "22"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "ip_range", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "rule_action", "Accept"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "priority", "50"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "remark", "ssh"),
					resource.TestCheckResourceAttrSet("ucloud_sec_group_rule.foo", "sec_group_id"),
				),
			},

			{
				// every field but sec_group_id is updated in place through
				// UpdateSecGroupRule, the rule id must stay the same
				Config: testAccSecGroupRuleConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &ruleSet),
					testAccCheckSecGroupRuleAttributes(&ruleSet),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "direction", "Ingress"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "protocol_type", "TCP"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "dst_port", "3306"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "ip_range", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "rule_action", "Drop"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "priority", "100"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "remark", "mysql"),
				),
			},
		},
	})
}

func testAccCheckSecGroupRuleExists(n string, ruleSet *vpc.SecGroupRuleInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("sec group rule id is empty")
		}

		secGroupId, ruleId, err := parseUCloudSecGroupRuleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeSecGroupRuleById(secGroupId, ruleId)

		log.Printf("[INFO] sec group rule id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*ruleSet = *ptr
		return nil
	}
}

func testAccCheckSecGroupRuleAttributes(ruleSet *vpc.SecGroupRuleInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ruleSet.RuleId == "" {
			return fmt.Errorf("sec group rule id is empty")
		}

		return nil
	}
}

func testAccCheckSecGroupRuleDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_sec_group_rule" {
			continue
		}

		secGroupId, ruleId, err := parseUCloudSecGroupRuleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeSecGroupRuleById(secGroupId, ruleId)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.RuleId != "" {
			return fmt.Errorf("sec group rule still exist")
		}
	}

	return nil
}

const testAccSecGroupRuleConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-rule"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-rule"
	vpc_id = "${ucloud_vpc.foo.id}"
}

resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = "${ucloud_sec_group.foo.id}"
	direction     = "Ingress"
	protocol_type = "TCP"
	dst_port      = "22"
	ip_range      = "192.168.0.0/16"
	rule_action   = "Accept"
	priority      = 50
	remark        = "ssh"
}
`

const testAccSecGroupRuleConfigTwo = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-rule"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-rule"
	vpc_id = "${ucloud_vpc.foo.id}"
}

resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = "${ucloud_sec_group.foo.id}"
	direction     = "Ingress"
	protocol_type = "TCP"
	dst_port      = "3306"
	ip_range      = "192.168.1.0/24"
	rule_action   = "Drop"
	priority      = 100
	remark        = "mysql"
}
`
