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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "name", "tf-acc-disk-basic"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "10"),
				),
			},

			{
				Config: testAccDiskConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "name", "tf-acc-disk-basic-update"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "20"),
				),
			},
		},
	})
}

func TestAccUCloudDisk_rssd(t *testing.T) {
	var diskSet udisk.UDiskDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDiskRssd,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "name", "tf-acc-disk-rssd"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "10"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_type", "rssd_data_disk"),
				),
			},
		},
	})
}

func TestAccUCloudDisk_tag(t *testing.T) {
	var diskSet udisk.UDiskDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDiskDefaultTag,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "name", "tf-acc-disk-tag"),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "10"),
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
			return fmt.Errorf("disk still exist")
		}
	}

	return nil
}

const testAccDiskConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-basic"
	tag               = "tf-acc"
	disk_size         = 10
}
`

const testAccDiskConfigUpdate = `
data "ucloud_zones" "default" {
}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-basic-update"
	tag               = "tf-acc"
	disk_size         = 20
}
`
const testAccDiskRssd = `
resource "ucloud_disk" "foo" {
	availability_zone = "cn-bj2-05"
	name              = "tf-acc-disk-rssd"
	tag               = "tf-acc"
	disk_size         = 10
	disk_type         = "rssd_data_disk"
}
`
const testAccDiskDefaultTag = `
locals {
	tag = ""
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-tag"
	tag               = "${local.tag}"
	disk_size         = 10
}
`
