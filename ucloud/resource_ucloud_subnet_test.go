package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudSubnet_basic(t *testing.T) {
	var val vpc.VPCSubnetInfoSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_subnet.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSubnetDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccSubnetConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("ucloud_subnet.foo", &val),
					testAccCheckSubnetAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_subnet.foo", "cidr_block", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("ucloud_subnet.foo", "name", "testAcc"),
					resource.TestCheckResourceAttr("ucloud_subnet.foo", "tag", "testTag"),
				),
			},

			resource.TestStep{
				Config: testAccSubnetConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("ucloud_subnet.foo", &val),
					testAccCheckSubnetAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_subnet.foo", "cidr_block", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("ucloud_subnet.foo", "name", "testAccTwo"),
					resource.TestCheckResourceAttr("ucloud_subnet.foo", "tag", "testTagTwo"),
				),
			},
		},
	})

}

func testAccCheckSubnetExists(n string, val *vpc.VPCSubnetInfoSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("subnet id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeSubnetById(rs.Primary.ID)

		log.Printf("[INFO] subnet id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckSubnetAttributes(val *vpc.VPCSubnetInfoSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.SubnetId == "" {
			return fmt.Errorf("subnet id is empty")
		}

		if val.VPCId == "" {
			return fmt.Errorf("vpc id has not been bound")
		}
		return nil
	}
}

func testAccCheckSubnetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_subnet" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeSubnetById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.SubnetId != "" {
			return fmt.Errorf("subnet still exist")
		}
	}

	return nil
}

const testAccSubnetConfig = `
resource "ucloud_vpc" "foo" {
	name = "testAccVPC"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name = "testAcc"
	tag = "testTag"
	cidr_block = "192.168.1.0/24"
	vpc_id = "${ucloud_vpc.foo.id}"
}
`
const testAccSubnetConfigTwo = `
resource "ucloud_vpc" "foo" {
	name = "testAccVPC"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name = "testAccTwo"
	tag = "testTagTwo"
	cidr_block = "192.168.1.0/24"
	vpc_id = "${ucloud_vpc.foo.id}"
}
`
