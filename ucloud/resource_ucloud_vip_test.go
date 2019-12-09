package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"log"
	"testing"
)

func TestAccUCloudVIP_basic(t *testing.T) {
	var vipSet unet.VIPDetailSet

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

func testAccCheckVIPExists(n string, vipSet *unet.VIPDetailSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vip id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVIPById(rs.Primary.ID)

		log.Printf("[INFO] vip id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*vipSet = *ptr
		return nil
	}
}

func testAccCheckVIPAttributes(vipSet *unet.VIPDetailSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if vipSet.VIPId == "" {
			return fmt.Errorf("vip id is empty")
		}
		return nil
	}
}

func testAccCheckVIPDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_vip" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVIPById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VIPId != "" {
			return fmt.Errorf("vip still exist")
		}
	}

	return nil
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
	vpc_id	 	= "${ucloud_vpc.foo.id}"
	subnet_id	= "${ucloud_subnet.foo.id}"
	name  	 	= "tf-acc-vip-basic"
	remark 		= "test"
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
	vpc_id	 	= "${ucloud_vpc.foo.id}"
	subnet_id	= "${ucloud_subnet.foo.id}"
	name  	 	= "tf-acc-vip-basic-update"
	remark 		= "test-update"
	tag         = "tf-acc"
}
`
