package vpc_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudSecGroupRule_basic(t *testing.T) {
	var val vpcapi.SecGroupRuleInfo
	var ruleID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_sec_group_rule.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecGroupRuleDestroy,

		Steps: []resource.TestStep{
			{
				// The ip range carries two CIDR blocks, and the group still has
				// to end up holding exactly one rule: the comma separated list
				// is a value inside one rule, not a way to add several.
				Config: testAccSecGroupRuleConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &val),
					testAccCheckSecGroupRuleAttributes(&val, "10.0.0.0/8,192.168.0.0/16"),
					testAccCheckSecGroupRuleCount("ucloud_sec_group.foo", 1),
					captureSecGroupRuleID(&val, &ruleID),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "protocol_type", "TCP"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "dst_port", "22"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "priority", "50"),
					resource.TestCheckResourceAttrPair("ucloud_sec_group_rule.foo", "sec_group_id", "ucloud_sec_group.foo", "id"),
				),
			},
			{
				// Changes the port and the priority. UpdateSecGroupRule addresses
				// the rule by its ID, so this is an edit rather than a
				// replacement, and the ID recorded above has to survive it.
				Config: testAccSecGroupRuleConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &val),
					testAccCheckSecGroupRuleIDSameAs(&val, &ruleID),
					testAccCheckSecGroupRuleCount("ucloud_sec_group.foo", 1),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "dst_port", "443,2000-10000"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "priority", "60"),
				),
			},
			{
				// Switches to a protocol a port carries no meaning for. dst_port
				// leaves the configuration, and whatever the API reports for it
				// has to be normalized away instead of diffing forever.
				Config: testAccSecGroupRuleConfigICMP,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &val),
					testAccCheckSecGroupRuleIDSameAs(&val, &ruleID),
					testAccCheckSecGroupRuleCount("ucloud_sec_group.foo", 1),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "protocol_type", "ICMP"),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "dst_port", ""),
				),
			},
		},
	})
}

// TestAccUCloudSecGroupRule_remarkCleared covers removing the remark from a rule
// that already carries one. It is the only attribute of the rule a configuration
// can take away rather than merely change, and the only one whose removal does
// not travel as a value: the update path builds its rule from the configuration,
// where the remark is the empty string, and the SDK's form encoder drops empty
// values before the request leaves, so UpdateSecGroupRule is never told to clear
// anything. Whether the API reads that absent parameter as "clear the remark" or
// as "leave it alone" is what this case settles.
//
// The two steps differ in the remark and nothing else, so a failure names it.
// What does the settling is the plan the framework runs after every step: if the
// API keeps the remark, the read writes it back, the plan is not empty, and the
// second step fails before its check is reached. The id is deliberately not
// asserted to be unchanged here - a rule replaced rather than edited would lose
// its remark without the API having cleared anything, which would answer the
// question wrongly.
func TestAccUCloudSecGroupRule_remarkCleared(t *testing.T) {
	var val vpcapi.SecGroupRuleInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_sec_group_rule.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecGroupRuleDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupRuleConfigRemark,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &val),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "remark", "ssh"),
				),
			},
			{
				Config: testAccSecGroupRuleConfigRemarkCleared,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupRuleExists("ucloud_sec_group_rule.foo", &val),
					resource.TestCheckResourceAttr("ucloud_sec_group_rule.foo", "remark", ""),
				),
			},
		},
	})
}

// captureSecGroupRuleID records the ID the API assigned, so a later step can
// tell an edit apart from a replacement.
func captureSecGroupRuleID(value *vpcapi.SecGroupRuleInfo, target *string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.RuleId == "" {
			return fmt.Errorf("sec group rule has no id")
		}
		*target = value.RuleId
		return nil
	}
}

// testAccCheckSecGroupRuleIDSameAs fails when the rule was replaced instead of
// edited. want is a pointer, like captureSecGroupRuleID's target, because the
// id is only known once the step that creates the rule has run: every Check in
// the Steps literal is built before the first step executes, so a plain string
// would be bound to the empty value ruleID still holds at that moment and the
// check could never pass.
func testAccCheckSecGroupRuleIDSameAs(value *vpcapi.SecGroupRuleInfo, want *string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if *want == "" {
			return fmt.Errorf("no sec group rule id was recorded by the previous step")
		}
		if value.RuleId != *want {
			return fmt.Errorf("sec group rule was replaced rather than edited: id = %q, want %q", value.RuleId, *want)
		}
		return nil
	}
}

const testAccSecGroupRuleConfigVPC = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-rule-vpc"
	tag         = ""
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-rule"
	vpc_id = ucloud_vpc.foo.id
}
`

const testAccSecGroupRuleConfig = testAccSecGroupRuleConfigVPC + `
resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = ucloud_sec_group.foo.id
	direction     = "Ingress"
	protocol_type = "TCP"
	dst_port      = "22"
	ip_range      = "10.0.0.0/8,192.168.0.0/16"
	rule_action   = "Accept"
	priority      = 50
	remark        = "ssh"
}
`

const testAccSecGroupRuleConfigUpdate = testAccSecGroupRuleConfigVPC + `
resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = ucloud_sec_group.foo.id
	direction     = "Ingress"
	protocol_type = "TCP"
	dst_port      = "443,2000-10000"
	ip_range      = "10.0.0.0/8,192.168.0.0/16"
	rule_action   = "Accept"
	priority      = 60
	remark        = "https"
}
`

const testAccSecGroupRuleConfigICMP = testAccSecGroupRuleConfigVPC + `
resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = ucloud_sec_group.foo.id
	direction     = "Egress"
	protocol_type = "ICMP"
	ip_range      = "0.0.0.0/0"
	rule_action   = "Accept"
	priority      = 100
}
`

const testAccSecGroupRuleConfigRemark = testAccSecGroupRuleConfigVPC + `
resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = ucloud_sec_group.foo.id
	direction     = "Ingress"
	protocol_type = "TCP"
	dst_port      = "22"
	ip_range      = "10.0.0.0/8"
	rule_action   = "Accept"
	priority      = 50
	remark        = "ssh"
}
`

// testAccSecGroupRuleConfigRemarkCleared is the configuration above with the
// remark removed and nothing else touched, so the remark is the only attribute
// the step between them has to settle.
const testAccSecGroupRuleConfigRemarkCleared = testAccSecGroupRuleConfigVPC + `
resource "ucloud_sec_group_rule" "foo" {
	sec_group_id  = ucloud_sec_group.foo.id
	direction     = "Ingress"
	protocol_type = "TCP"
	dst_port      = "22"
	ip_range      = "10.0.0.0/8"
	rule_action   = "Accept"
	priority      = 50
}
`
