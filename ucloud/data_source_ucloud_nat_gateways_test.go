package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudNatGatewaysDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataNatGatewaysConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_nat_gateways.foo"),
					resource.TestCheckResourceAttr("data.ucloud_nat_gateways.foo", "nat_gateways.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_nat_gateways.foo", "nat_gateways.0.name", "tf-acc-nat-gateways"),
					resource.TestCheckResourceAttr("data.ucloud_nat_gateways.foo", "nat_gateways.0.tag", "tf-acc"),
				),
			},
		},
	})
}

const testAccDataNatGatewaysConfig = `
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-nat-gateways"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-nat-gateways"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_subnet" "bar" {
  name       = "tf-acc-nat-gateways"
  tag        = "tf-acc"
  cidr_block = "192.168.2.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-nat-gateways"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
  type = "recommend_web"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id, ucloud_subnet.bar.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateways"
  tag               = "tf-acc"
  enable_white_list = false
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}

data "ucloud_nat_gateways" "foo" {
	ids = ucloud_nat_gateway.foo.*.id
}
`
