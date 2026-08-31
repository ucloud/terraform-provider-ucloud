package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudVIP_basic(t *testing.T) {
	var vipSet vpcapi.VIPDetailSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vip.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVIPDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVIPConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVIPExists("ucloud_vip.foo", &vipSet),
					testAccCheckVIPAttributes(&vipSet),
					resource.TestCheckResourceAttr("ucloud_vip.foo", "name", "tf-acc-vip-basic"),
					resource.TestCheckResourceAttr("ucloud_vip.foo", "remark", "test"),
				),
			},
			{
				Config: testAccVIPConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVIPExists("ucloud_vip.foo", &vipSet),
					testAccCheckVIPAttributes(&vipSet),
					resource.TestCheckResourceAttr("ucloud_vip.foo", "name", "tf-acc-vip-basic-update"),
					resource.TestCheckResourceAttr("ucloud_vip.foo", "remark", "test-update"),
				),
			},
		},
	})
}

const testAccVIPConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vip"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-vip"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}
resource "ucloud_vip" "foo" {
	vpc_id      = "${ucloud_vpc.foo.id}"
	subnet_id   = "${ucloud_subnet.foo.id}"
	name        = "tf-acc-vip-basic"
	remark      = "test"
	tag         = "tf-acc"
}
`

const testAccVIPConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vip"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-vip"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}
resource "ucloud_vip" "foo" {
	vpc_id      = "${ucloud_vpc.foo.id}"
	subnet_id   = "${ucloud_subnet.foo.id}"
	name        = "tf-acc-vip-basic-update"
	remark      = "test-update"
	tag         = "tf-acc"
}
`
