package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUCloudVPNCusGW_basic(t *testing.T) {
	var val ipsecvpn.RemoteVPNGatewayDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpn_customer_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPNCusGWDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPNCusGWConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNCusGWExists("ucloud_vpn_customer_gateway.foo", &val),
					testAccCheckVPNCusGWAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpn_customer_gateway.foo", "name", "tf-acc-vpn-customer-gateway-basic"),
				),
			},
		},
	})
}

func testAccCheckVPNCusGWExists(n string, val *ipsecvpn.RemoteVPNGatewayDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vpn customer gateway id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVPNCustomerGatewayById(rs.Primary.ID)

		log.Printf("[INFO] vpn customer gateway id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckVPNCusGWAttributes(val *ipsecvpn.RemoteVPNGatewayDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.RemoteVPNGatewayId == "" {
			return fmt.Errorf("vpn customer gateway id is empty")
		}

		return nil
	}
}

func testAccCheckVPNCusGWDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_vpn_customer_gateway" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVPNCustomerGatewayById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.RemoteVPNGatewayId != "" {
			return fmt.Errorf("vpn customer gateway still exist")
		}
	}

	return nil
}

const testAccVPNCusGWConfig = `
resource "ucloud_vpn_customer_gateway" "foo" {
    ip_address  = "10.0.0.1"
	name 		= "tf-acc-vpn-customer-gateway-basic"
	tag         = "tf-acc"
}
`
