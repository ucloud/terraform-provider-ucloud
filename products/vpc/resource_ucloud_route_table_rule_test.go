package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudRouteTableRule_basic(t *testing.T) {
	var rule vpcapi.RouteRuleInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_route_table_rule.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableRuleDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableRuleConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableRuleExists("ucloud_route_table_rule.foo", &rule),
					testAccCheckRouteTableRuleAttributes(&rule),
					resource.TestCheckResourceAttr("ucloud_route_table_rule.foo", "dst_addr", "10.8.0.0/16"),
					resource.TestCheckResourceAttr("ucloud_route_table_rule.foo", "nexthop_type", "vip"),
					resource.TestCheckResourceAttr("ucloud_route_table_rule.foo", "remark", "tf-acc-route-table-rule"),
				),
			},
			{
				Config: testAccRouteTableRuleConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableRuleExists("ucloud_route_table_rule.foo", &rule),
					testAccCheckRouteTableRuleAttributes(&rule),
					resource.TestCheckResourceAttr("ucloud_route_table_rule.foo", "dst_addr", "10.8.0.0/16"),
					resource.TestCheckResourceAttr("ucloud_route_table_rule.foo", "nexthop_type", "vip"),
					resource.TestCheckResourceAttr("ucloud_route_table_rule.foo", "remark", "tf-acc-route-table-rule-updated"),
				),
			},
		},
	})
}

const testAccRouteTableRuleConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-route-table-rule"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-route-table-rule"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_vip" "foo" {
	vpc_id    = ucloud_vpc.foo.id
	subnet_id = ucloud_subnet.foo.id
	name      = "tf-acc-route-table-rule"
	tag       = "tf-acc"
}

resource "ucloud_route_table" "foo" {
	name   = "tf-acc-route-table-rule"
	tag    = "tf-acc"
	vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table_rule" "foo" {
	route_table_id = ucloud_route_table.foo.id
	dst_addr       = "10.8.0.0/16"
	nexthop_type   = "vip"
	nexthop_id     = ucloud_vip.foo.id
	remark         = "tf-acc-route-table-rule"
}
`

const testAccRouteTableRuleConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-route-table-rule"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-route-table-rule"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_vip" "foo" {
	vpc_id    = ucloud_vpc.foo.id
	subnet_id = ucloud_subnet.foo.id
	name      = "tf-acc-route-table-rule"
	tag       = "tf-acc"
}

resource "ucloud_route_table" "foo" {
	name   = "tf-acc-route-table-rule"
	tag    = "tf-acc"
	vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table_rule" "foo" {
	route_table_id = ucloud_route_table.foo.id
	dst_addr       = "10.8.0.0/16"
	nexthop_type   = "vip"
	nexthop_id     = ucloud_vip.foo.id
	remark         = "tf-acc-route-table-rule-updated"
}
`
