package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudVPNCustomerGatewaysDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataVPNCustomerGatewaysConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_vpn_customer_gateways.foo"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_customer_gateways.foo", "vpn_customer_gateways.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_customer_gateways.foo", "vpn_customer_gateways.0.ip_address", "10.0.0.1"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_customer_gateways.foo", "vpn_customer_gateways.0.name", "tf-acc-vpn-customer-gateways"),
				),
			},
		},
	})
}

const testAccDataVPNCustomerGatewaysConfig = `
resource "ucloud_vpn_customer_gateway" "foo" {
    ip_address  = "10.0.0.1"
	name 		= "tf-acc-vpn-customer-gateways"
	tag         = "tf-acc"
}

data "ucloud_vpn_customer_gateways" "foo" {
	ids = ucloud_vpn_customer_gateway.foo.*.id
}
`
