package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudVPC_basic(t *testing.T) {
	var val vpcapi.VPCInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpc.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPCDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists("ucloud_vpc.foo", &val),
					testAccCheckVPCAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "name", "tf-acc-vpc"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.3901788224", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.#", "1"),
				),
			},
		},
	})
}

func TestAccUCloudVPC_update_network(t *testing.T) {
	var val vpcapi.VPCInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpc.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPCDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists("ucloud_vpc.foo", &val),
					testAccCheckVPCAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "name", "tf-acc-vpc"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.3901788224", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.#", "1"),
				),
			},
			{
				Config: testAccVPCConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists("ucloud_vpc.foo", &val),
					testAccCheckVPCAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "name", "tf-acc-vpc"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.#", "2"),
				),
			},
			{
				Config: testAccVPCConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists("ucloud_vpc.foo", &val),
					testAccCheckVPCAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "name", "tf-acc-vpc"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.3901788224", "192.168.0.0/16"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.#", "1"),
				),
			},
		},
	})
}

const testAccVPCConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpc"
	tag         = ""
	cidr_blocks = ["192.168.0.0/16"]
}
`

const testAccVPCConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpc"
	tag         = ""
	cidr_blocks = [
		"192.168.0.0/16",
		"172.16.0.0/16"
	]
}
`
