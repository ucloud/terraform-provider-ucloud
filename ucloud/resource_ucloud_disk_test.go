package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
)

func TestAccUCloudDisk_basic(t *testing.T) {
	var diskSet udisk.UDiskDataSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDiskConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "name", "testAcc"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "10"),
				),
			},

			resource.TestStep{
				Config: testAccDiskConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "name", "testAccTwo"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "20"),
				),
			},
		},
	})

}

func testAccCheckDiskExists(n string, diskSet *udisk.UDiskDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("disk id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeDiskById(rs.Primary.ID)

		log.Printf("[INFO] disk id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*diskSet = *ptr
		return nil
	}
}

func testAccCheckDiskAttributes(diskSet *udisk.UDiskDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if diskSet.UDiskId == "" {
			return fmt.Errorf("disk id is empty")
		}
		return nil
	}
}

func testAccCheckDiskDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_disk" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeDiskById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.UDiskId != "" {
			return fmt.Errorf("Disk still exist")
		}
	}

	return nil
}

const testAccDiskConfig = `
resource "ucloud_disk" "foo" {
	availability_zone = "cn-sh2-02"
	name = "testAcc"
	disk_size = 10
}
`
const testAccDiskConfigTwo = `
resource "ucloud_disk" "foo" {
	availability_zone = "cn-sh2-02"
	name = "testAccTwo"
	disk_size = 20
}
`
