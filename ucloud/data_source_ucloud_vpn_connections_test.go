package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudVPNConnectionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataVPNConnectionsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_vpn_connections.foo"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.name", "tf-acc-vpn-connections"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.ike_config.0.exchange_mode", "aggressive"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.ike_config.0.sa_life_time", "10000"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.ipsec_config.0.encryption_algorithm", "aes192"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.ipsec_config.0.sa_life_time", "20000"),
					resource.TestCheckResourceAttr("data.ucloud_vpn_connections.foo", "vpn_connections.0.ipsec_config.0.sa_life_time_bytes", "200000"),
				),
			},
		},
	})
}

const testAccDataVPNConnectionsConfig = `
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-vpn-connections"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-vpn-connections"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-vpn-connections"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
  vpc_id = ucloud_vpc.foo.id
  grade  = "standard"
  eip_id = ucloud_eip.foo.id
  name   = "tf-acc-vpn-connections"
  tag    = "tf-acc"
}

resource "ucloud_vpn_customer_gateway" "foo" {
  ip_address = "10.0.0.1"
  name       = "tf-acc-vpn-connections"
  tag        = "tf-acc"
}

resource "ucloud_vpn_connection" "foo" {
  vpn_gateway_id      = "${ucloud_vpn_gateway.foo.id}"
  customer_gateway_id = "${ucloud_vpn_customer_gateway.foo.id}"
  vpc_id              = "${ucloud_vpc.foo.id}"
  name                = "tf-acc-vpn-connections"
  tag                 = "tf-acc"
  remark              = "test"
  ike_config {
	ike_version              = "ikev1"
    pre_shared_key           = "test_password_1"
    exchange_mode            = "aggressive"
    encryption_algorithm     = "aes192"
    authentication_algorithm = "md5"
    local_id                 = "auto"
    remote_id                = "auto"
    dh_group                 = 14
    sa_life_time             = 10000
  }
  ipsec_config {
    local_subnet_ids         = ["${ucloud_subnet.foo.id}"]
    remote_subnets           = ["10.0.0.0/24"]
    protocol                 = "ah"
    encryption_algorithm     = "aes192"
    authentication_algorithm = "md5"
    sa_life_time             = 20000
	sa_life_time_bytes		 = 200000
    pfs_dh_group			 = "disable"
  }
}

data "ucloud_vpn_connections" "foo" {
	ids = ucloud_vpn_connection.foo.*.id
}
`
