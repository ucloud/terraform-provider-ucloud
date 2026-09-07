package ipsecvpn_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
)

func TestAccUCloudVPNConn_basic(t *testing.T) {
	var value ipsecvpn.VPNTunnelDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpn_connection.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPNConnectionDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConnectionExists("ucloud_vpn_connection.foo", &value),
					testAccCheckVPNConnectionAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "name", "tf-acc-vpn-connection-basic"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.#", "1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.#", "1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.ike_version", "ikev1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.pre_shared_key", "test_password_1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.exchange_mode", "aggressive"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.encryption_algorithm", "aes192"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.authentication_algorithm", "md5"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.local_id", "auto"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.remote_id", "auto"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.dh_group", "14"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.sa_life_time", "10000"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.protocol", "ah"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.encryption_algorithm", "aes192"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.authentication_algorithm", "md5"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.sa_life_time", "20000"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.sa_life_time_bytes", "200000"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.pfs_dh_group", "disable"),
				),
			},

			{
				Config: testAccVPNConnConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPNConnectionExists("ucloud_vpn_connection.foo", &value),
					testAccCheckVPNConnectionAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "name", "tf-acc-vpn-connection-basic"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.#", "1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.#", "1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.ike_version", "ikev1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.pre_shared_key", "test_password_2"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.exchange_mode", "main"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.encryption_algorithm", "aes256"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.authentication_algorithm", "sha2-256"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.local_id", "auto"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.remote_id", "auto"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.dh_group", "16"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ike_config.0.sa_life_time", "30000"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.protocol", "esp"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.encryption_algorithm", "aes128"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.authentication_algorithm", "sha1"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.sa_life_time", "40000"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.sa_life_time_bytes", "500000"),
					resource.TestCheckResourceAttr("ucloud_vpn_connection.foo", "ipsec_config.0.pfs_dh_group", "5"),
				),
			},
		},
	})
}

func testAccCheckVPNConnectionExists(name string, value *ipsecvpn.VPNTunnelDataSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("vpn connection id is empty")
		}

		client, err := testAccIPSecVPNClient()
		if err != nil {
			return err
		}
		pointer, found, err := describeAccVPNConnectionByID(client, item.Primary.ID)
		log.Printf("[INFO] vpn connection id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("vpn connection %q is not found", item.Primary.ID)
		}
		*value = *pointer
		return nil
	}
}

func testAccCheckVPNConnectionAttributes(value *ipsecvpn.VPNTunnelDataSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.VPNTunnelId == "" {
			return fmt.Errorf("vpn connection id is empty")
		}
		return nil
	}
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
  vpc_id = ucloud_vpc.foo.id
  grade  = "standard"
  eip_id = ucloud_eip.foo.id
  name   = "tf-acc-vpn-connection-basic"
  tag    = "tf-acc"
}

resource "ucloud_vpn_customer_gateway" "foo" {
  ip_address = "10.0.0.1"
  name       = "tf-acc-vpn-connection-basic"
  tag        = "tf-acc"
}

resource "ucloud_vpn_connection" "foo" {
  vpn_gateway_id      = "${ucloud_vpn_gateway.foo.id}"
  customer_gateway_id = "${ucloud_vpn_customer_gateway.foo.id}"
  vpc_id              = "${ucloud_vpc.foo.id}"
  name                = "tf-acc-vpn-connection-basic"
  tag                 = "tf-acc"
  remark              = "test"
  ike_config {
    ike_version              = "ikev1"
    pre_shared_key           = "test_password_1"
    exchange_mode            = "aggressive"
    encryption_algorithm     = "aes192"
    authentication_algorithm = "md5"
    local_id                = "auto"
    remote_id               = "auto"
    dh_group                = 14
    sa_life_time            = 10000
  }
  ipsec_config {
    local_subnet_ids         = ["${ucloud_subnet.foo.id}"]
    remote_subnets           = ["10.0.0.0/24"]
    protocol                 = "ah"
    encryption_algorithm     = "aes192"
    authentication_algorithm = "md5"
    sa_life_time             = 20000
    sa_life_time_bytes       = 200000
    pfs_dh_group             = "disable"
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
  vpc_id = ucloud_vpc.foo.id
  grade  = "standard"
  eip_id = ucloud_eip.foo.id
  name   = "tf-acc-vpn-connection-basic"
  tag    = "tf-acc"
}

resource "ucloud_vpn_customer_gateway" "foo" {
  ip_address = "10.0.0.1"
  name       = "tf-acc-vpn-connection-basic"
  tag        = "tf-acc"
}

resource "ucloud_vpn_connection" "foo" {
  vpn_gateway_id      = "${ucloud_vpn_gateway.foo.id}"
  customer_gateway_id = "${ucloud_vpn_customer_gateway.foo.id}"
  vpc_id              = "${ucloud_vpc.foo.id}"
  name                = "tf-acc-vpn-connection-basic"
  tag                 = "tf-acc"
  remark              = "test"
  ike_config {
    pre_shared_key           = "test_password_2"
    exchange_mode            = "main"
    encryption_algorithm     = "aes256"
    authentication_algorithm = "sha2-256"
    local_id                = "auto"
    remote_id               = "auto"
    dh_group                = 16
    sa_life_time            = 30000
  }
  ipsec_config {
    local_subnet_ids         = ["${ucloud_subnet.foo.id}"]
    remote_subnets           = ["10.0.0.0/24"]
    protocol                 = "esp"
    encryption_algorithm     = "aes128"
    authentication_algorithm = "sha1"
    sa_life_time             = 40000
    sa_life_time_bytes       = 500000
    pfs_dh_group             = "5"
  }
}
`
