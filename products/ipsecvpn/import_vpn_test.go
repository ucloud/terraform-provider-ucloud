package ipsecvpn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudVPNGateway_import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPNGatewayDestroy,
		Steps: []resource.TestStep{
			{Config: testAccVPNGWConfig},
			{
				ResourceName:      "ucloud_vpn_gateway.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUCloudVPNCustomerGateway_import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPNCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{Config: testAccVPNCusGWConfig},
			{
				ResourceName:      "ucloud_vpn_customer_gateway.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUCloudVPNConnection_import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{Config: testAccVPNConnConfig},
			{
				ResourceName:      "ucloud_vpn_connection.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
