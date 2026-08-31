package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudNatGWRule_basic(t *testing.T) {
	var natGWSet vpcapi.NatGatewayDataSet
	var ruleSet vpcapi.NATGWPolicyDataSet
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
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
  enable_white_list = false
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}

resource "ucloud_nat_gateway_rule" "foo" {
  nat_gateway_id = ucloud_nat_gateway.foo.id
  protocol       = "tcp"
  src_eip_id     = ucloud_eip.foo.id
  src_port_range = "90-100"
  dst_ip         = ucloud_instance.foo.private_ip
  dst_port_range = "90-100"
  name           = "tf-acc-nat-gateway-rule-basic"
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
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
  enable_white_list = false
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}

resource "ucloud_nat_gateway_rule" "foo" {
  nat_gateway_id = ucloud_nat_gateway.foo.id
  protocol       = "tcp"
  src_eip_id     = ucloud_eip.foo.id
  src_port_range = "100-110"
  dst_ip         = ucloud_instance.foo.private_ip
  dst_port_range = "100-110"
  name           = "tf-acc-nat-gateway-rule-update"
}
`
