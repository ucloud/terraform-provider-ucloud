package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudSecGroup_basic(t *testing.T) {
	var sgSet vpc.SecGroupInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_sec_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &sgSet),
					testAccCheckSecGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "name", "tf-acc-sec-group"),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "remark", "acceptance test"),
					resource.TestCheckResourceAttrSet("ucloud_sec_group.foo", "vpc_id"),
					resource.TestCheckResourceAttrSet("ucloud_sec_group.foo", "create_time"),
				),
			},

			{
				Config: testAccSecGroupConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &sgSet),
					testAccCheckSecGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "name", "tf-acc-sec-group-two"),
					// tag is omitted by the second config, it must fall back to the default
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "remark", "acceptance test updated"),
				),
			},
		},
	})
}

func testAccCheckSecGroupExists(n string, sgSet *vpc.SecGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("sec group id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeSecGroupById(rs.Primary.ID)

		log.Printf("[INFO] sec group id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*sgSet = *ptr
		return nil
	}
}

func testAccCheckSecGroupAttributes(sgSet *vpc.SecGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sgSet.SecGroupId == "" {
			return fmt.Errorf("sec group id is empty")
		}

		if sgSet.VPCId == "" {
			return fmt.Errorf("sec group %q is expected to belong to a vpc", sgSet.SecGroupId)
		}

		return nil
	}
}

func testAccCheckSecGroupDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_sec_group" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeSecGroupById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.SecGroupId != "" {
			return fmt.Errorf("sec group still exist")
		}
	}

	return nil
}

const testAccSecGroupConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group"
	vpc_id = "${ucloud_vpc.foo.id}"
	tag    = "tf-acc"
	remark = "acceptance test"
}
`

const testAccSecGroupConfigTwo = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-two"
	vpc_id = "${ucloud_vpc.foo.id}"
	remark = "acceptance test updated"
}
`
