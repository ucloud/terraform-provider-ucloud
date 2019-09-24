package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUCloudNatGWRule_basic(t *testing.T) {
	var natGWSet vpc.NatGatewayDataSet
	var ruleSet vpc.NATGWPolicyDataSet
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_nat_gateway_rule.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNatGWRuleDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccNatGWRuleConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGWExists("ucloud_nat_gateway.foo", &natGWSet),
					testAccCheckNatGWRuleExists("ucloud_nat_gateway_rule.foo", &natGWSet, &ruleSet),
					testAccCheckNatGWRuleAttributes(&ruleSet),
					resource.TestCheckResourceAttr("ucloud_nat_gateway_rule.foo", "name", "tf-acc-nat-gateway-rule-basic"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway_rule.foo", "src_port_range", "90-100"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway_rule.foo", "dst_port_range", "90-100"),
				),
			},

			{
				Config: testAccNatGWRuleConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGWExists("ucloud_nat_gateway.foo", &natGWSet),
					testAccCheckNatGWRuleExists("ucloud_nat_gateway_rule.foo", &natGWSet, &ruleSet),
					testAccCheckNatGWRuleAttributes(&ruleSet),
					resource.TestCheckResourceAttr("ucloud_nat_gateway_rule.foo", "name", "tf-acc-nat-gateway-rule-update"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway_rule.foo", "src_port_range", "100-110"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway_rule.foo", "dst_port_range", "100-110"),
				),
			},
		},
	})
}

func testAccCheckNatGWRuleExists(n string, natGWSet *vpc.NatGatewayDataSet, ruleSet *vpc.NATGWPolicyDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("nat_gateway rule id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeNatGatewayRuleById(rs.Primary.ID, natGWSet.NATGWId)

		log.Printf("[INFO] nat_gateway rule id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*ruleSet = *ptr
		return nil
	}
}

func testAccCheckNatGWRuleAttributes(ruleSet *vpc.NATGWPolicyDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ruleSet.PolicyId == "" {
			return fmt.Errorf("nat_gateway rule id is empty")
		}
		return nil
	}
}

func testAccCheckNatGWRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_nat_gateway_rule" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeNatGatewayRuleById(
			rs.Primary.ID,
			rs.Primary.Attributes["nat_gateway_id"],
		)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.PolicyId != "" {
			return fmt.Errorf("nat_gateway rule still exist")
		}
	}

	return nil
}

const testAccNatGWRuleConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-nat-gateway-rule-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-nat-gateway-rule-basic"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-nat-gateway-rule-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
	type = "recommend_web"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  vpc_id	 	    = ucloud_vpc.foo.id
  subnet_id	 	    = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
	vpc_id	 	   = ucloud_vpc.foo.id
	subnet_ids	   = [ucloud_subnet.foo.id]
	eip_id		   = ucloud_eip.foo.id
	name 		   = "tf-acc-nat-gateway-rule-basic"
	tag            = "tf-acc"
	security_group = data.ucloud_security_groups.foo.security_groups.0.id
}

resource "ucloud_nat_gateway_rule" "foo" {
	nat_gateway_id = ucloud_nat_gateway.foo.id
	protocol      =  "tcp"
	src_eip_id 	  = ucloud_eip.foo.id
	src_port_range = "90-100"
	dst_ip		   = ucloud_instance.foo.private_ip
	dst_port_range = "90-100"
	name 		   = "tf-acc-nat-gateway-rule-basic"
}
`

const testAccNatGWRuleConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-nat-gateway-rule-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-nat-gateway-rule-basic"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-nat-gateway-rule-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
	type = "recommend_web"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  vpc_id	 	    = ucloud_vpc.foo.id
  subnet_id	 	    = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
	vpc_id	 	   = ucloud_vpc.foo.id
	subnet_ids	   = [ucloud_subnet.foo.id]
	eip_id		   = ucloud_eip.foo.id
	name 		   = "tf-acc-nat-gateway-rule-basic"
	tag            = "tf-acc"
	security_group = data.ucloud_security_groups.foo.security_groups.0.id
}

resource "ucloud_nat_gateway_rule" "foo" {
	nat_gateway_id = ucloud_nat_gateway.foo.id
	protocol      =  "tcp"
	src_eip_id 	  = ucloud_eip.foo.id
	src_port_range = "100-110"
	dst_ip		   = ucloud_instance.foo.private_ip
	dst_port_range = "100-110"
	name 		   = "tf-acc-nat-gateway-rule-update"
}
`
