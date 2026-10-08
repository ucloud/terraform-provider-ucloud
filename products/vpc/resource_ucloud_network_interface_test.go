package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudNetworkInterface_basic(t *testing.T) {
	var uni vpcapi.NetworkInterface

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_network_interface.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNetworkInterfaceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfaceExists("ucloud_network_interface.foo", &uni),
					testAccCheckNetworkInterfaceAttributes(&uni),
					resource.TestCheckResourceAttrPair("ucloud_network_interface.foo", "vpc_id", "ucloud_vpc.foo", "id"),
					resource.TestCheckResourceAttrPair("ucloud_network_interface.foo", "subnet_id", "ucloud_subnet.foo", "id"),
					resource.TestCheckResourceAttr("ucloud_network_interface.foo", "name", "tf-acc-uni"),
				),
			},
			{
				Config: testAccNetworkInterfaceConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkInterfaceExists("ucloud_network_interface.foo", &uni),
					testAccCheckNetworkInterfaceAttributes(&uni),
					resource.TestCheckResourceAttr("ucloud_network_interface.foo", "name", "tf-acc-uni-updated"),
					resource.TestCheckResourceAttr("ucloud_network_interface.foo", "remark", "updated"),
				),
			},
		},
	})
}

const testAccNetworkInterfaceConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-uni"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-uni"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_network_interface" "foo" {
	name      = "tf-acc-uni"
	tag       = "tf-acc"
	remark    = "created"
	vpc_id    = ucloud_vpc.foo.id
	subnet_id = ucloud_subnet.foo.id
}
`

const testAccNetworkInterfaceConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-uni"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-uni"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_network_interface" "foo" {
	name      = "tf-acc-uni-updated"
	tag       = "tf-acc"
	remark    = "updated"
	vpc_id    = ucloud_vpc.foo.id
	subnet_id = ucloud_subnet.foo.id
}
`
