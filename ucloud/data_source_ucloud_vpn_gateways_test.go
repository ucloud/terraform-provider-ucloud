package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudVPNGatewaysDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataVPNGatewaysConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_vpn_gateways.foo"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_gateways.foo", "vpn_gateways.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_gateways.foo", "vpn_gateways.0.grade", "standard"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_gateways.foo", "vpn_gateways.0.name", "tf-acc-vpn-gateways"),
				),
			},
		},
	})
}

const testAccDataVPNGatewaysConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpn-gateways"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-vpn-gateways"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	grade		= "standard"
	eip_id		= ucloud_eip.foo.id
	name 		= "tf-acc-vpn-gateways"
	tag         = "tf-acc"
}

data "ucloud_vpn_gateways" "foo" {
	ids = ucloud_vpn_gateway.foo.*.id
}
`
