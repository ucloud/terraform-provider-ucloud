package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccUCloudVPNGW_basic(t *testing.T) {
	var val ipsecvpn.VPNGatewayDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpn_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPNGWDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPNGWConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGWExists("ucloud_vpn_gateway.foo", &val),
					testAccCheckVPNGWAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpn_gateway.foo", "name", "tf-acc-vpn-gateway-basic"),
					resource.TestCheckResourceAttr("ucloud_vpn_gateway.foo", "grade", "standard"),
				),
			},

			{
				Config: testAccVPNGWConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNGWExists("ucloud_vpn_gateway.foo", &val),
					testAccCheckVPNGWAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpn_gateway.foo", "name", "tf-acc-vpn-gateway-basic"),
					resource.TestCheckResourceAttr("ucloud_vpn_gateway.foo", "grade", "enhanced"),
				),
			},
		},
	})
}

func testAccCheckVPNGWExists(n string, val *ipsecvpn.VPNGatewayDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vpn gateway id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVPNGatewayById(rs.Primary.ID)

		log.Printf("[INFO] vpn gateway id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckVPNGWAttributes(val *ipsecvpn.VPNGatewayDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.VPNGatewayId == "" {
			return fmt.Errorf("vpn gateway id is empty")
		}

		return nil
	}
}

func testAccCheckVPNGWDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_vpn_gateway" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVPNGatewayById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VPNGatewayId != "" {
			return fmt.Errorf("vpn gateway still exist")
		}
	}

	return nil
}

const testAccVPNGWConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpn-gateway-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-vpn-gateway-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	grade		= "standard"
	eip_id		= ucloud_eip.foo.id
	name 		= "tf-acc-vpn-gateway-basic"
	tag         = "tf-acc"
}
`

const testAccVPNGWConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpn-gateway-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-vpn-gateway-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	grade		= "enhanced"
	eip_id		= ucloud_eip.foo.id
	name 		= "tf-acc-vpn-gateway-basic"
	tag         = "tf-acc"
}
`
