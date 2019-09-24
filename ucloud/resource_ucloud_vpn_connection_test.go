package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUCloudVPNConn_basic(t *testing.T) {
	var val ipsecvpn.VPNTunnelDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpn_connection.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPNConnDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConnExists("ucloud_vpn_connection.foo", &val),
					testAccCheckVPNConnAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "name", "tf-acc-vpn-connection-basic"),
				),
			},

			{
				Config: testAccVPNConnConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConnExists("ucloud_vpn_connection.foo", &val),
					testAccCheckVPNConnAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "name", "tf-acc-vpn-connection-basic"),
				),
			},
		},
	})
}

func testAccCheckVPNConnExists(n string, val *ipsecvpn.VPNTunnelDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vpn connection id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVPNConnectionById(rs.Primary.ID)

		log.Printf("[INFO] vpn connection id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckVPNConnAttributes(val *ipsecvpn.VPNTunnelDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.VPNTunnelId == "" {
			return fmt.Errorf("vpn connection id is empty")
		}

		return nil
	}
}

func testAccCheckVPNConnDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_vpn_connection" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVPNConnectionById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VPNTunnelId != "" {
			return fmt.Errorf("vpn connection still exist")
		}
	}

	return nil
}

const testAccVPNConnConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpn-connection-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-vpn-connection-basic"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-vpn-connection-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	grade		= "standard"
	eip_id		= ucloud_eip.foo.id
	name 		= "tf-acc-vpn-connection-basic"
	tag         = "tf-acc"
}

resource "ucloud_vpn_customer_gateway" "foo" {
    ip_address  = "10.0.0.1"
	name 		= "tf-acc-vpn-connection-basic"
	tag         = "tf-acc"
}

resource "ucloud_vpn_connection" "foo" {
    vpn_gateway_id  = "${ucloud_vpn_gateway.foo.id}"
	customer_gateway_id = "${ucloud_vpn_customer_gateway.foo.id}"
	vpc_id			= "${ucloud_vpc.foo.id}"
	name 			= "tf-acc-vpn-connection-basic"
	tag         	= "tf-acc"
	remark          = "test"
	ike_config {
		pre_shared_key = "123456"
		exchange_mode = "aggressive"
		encryption_algorithm = "aes192"
		authentication_algorithm = "md5"
		local_id	= "auto"
		remote_id   = "auto"
		dh_group	= 14
		sa_life_time = 10000
	}
	ipsec_config {
		local_subnet_ids = ["${ucloud_subnet.foo.id}"]
		remote_subnets = ["10.0.0.0/24"]
		protocol     = "ah"
		encryption_algorithm = "aes192"
		authentication_algorithm = "md5"
		sa_life_time  = 10000
	}
}
`
const testAccVPNConnConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpn-connection-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-vpn-connection-basic"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-vpn-connection-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	grade		= "standard"
	eip_id		= ucloud_eip.foo.id
	name 		= "tf-acc-vpn-connection-basic"
	tag         = "tf-acc"
}

resource "ucloud_vpn_customer_gateway" "foo" {
    ip_address  = "10.0.0.1"
	name 		= "tf-acc-vpn-connection-basic"
	tag         = "tf-acc"
}

resource "ucloud_vpn_connection" "foo" {
    vpn_gateway_id  = "${ucloud_vpn_gateway.foo.id}"
	customer_gateway_id = "${ucloud_vpn_customer_gateway.foo.id}"
	vpc_id			= "${ucloud_vpc.foo.id}"
	name 			= "tf-acc-vpn-connection-basic"
	tag         	= "tf-acc"
	remark          = "test"
	ike_config {
		pre_shared_key = "1234567"
		exchange_mode = "main"
		encryption_algorithm = "aes256"
		authentication_algorithm = "sha2-256"
		local_id	= "auto"
		remote_id   = "auto"
		dh_group	= 16
		sa_life_time = 20000
	}
	ipsec_config {
		local_subnet_ids = ["${ucloud_subnet.foo.id}"]
		remote_subnets = ["10.0.0.0/24"]
		protocol     = "esp"
		encryption_algorithm = "aes128"
		authentication_algorithm = "sha1"
		sa_life_time  = 30000
		sa_life_time_bytes = 9000
	}
}
`
