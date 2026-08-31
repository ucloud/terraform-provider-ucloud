package ipsecvpn_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
)

func TestAccUCloudVPNCusGW_basic(t *testing.T) {
	var value ipsecvpn.RemoteVPNGatewayDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpn_customer_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPNCustomerGatewayDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPNCusGWConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNCustomerGatewayExists("ucloud_vpn_customer_gateway.foo", &value),
					testAccCheckVPNCustomerGatewayAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_vpn_customer_gateway.foo", "name", "tf-acc-vpn-customer-gateway-basic"),
					resource.TestCheckResourceAttr("ucloud_vpn_customer_gateway.foo", "ip_address", "10.0.0.1"),
				),
			},
		},
	})
}

func testAccCheckVPNCustomerGatewayExists(
	name string,
	value *ipsecvpn.RemoteVPNGatewayDataSet,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("vpn customer gateway id is empty")
		}

		client, err := testAccIPSecVPNClient()
		if err != nil {
			return err
		}
		pointer, found, err := describeAccVPNCustomerGatewayByID(client, item.Primary.ID)
		log.Printf("[INFO] vpn customer gateway id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("vpn customer gateway %q is not found", item.Primary.ID)
		}
		*value = *pointer
		return nil
	}
}

func testAccCheckVPNCustomerGatewayAttributes(value *ipsecvpn.RemoteVPNGatewayDataSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.RemoteVPNGatewayId == "" {
			return fmt.Errorf("vpn customer gateway id is empty")
		}
		return nil
	}
}

const testAccVPNCusGWConfig = `
resource "ucloud_vpn_customer_gateway" "foo" {
	ip_address = "10.0.0.1"
	name       = "tf-acc-vpn-customer-gateway-basic"
	tag        = "tf-acc"
}
`
