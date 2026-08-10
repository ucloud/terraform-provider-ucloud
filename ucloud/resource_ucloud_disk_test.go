package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

// TestAccUCloudDisk_fromSnapshot verifies that a disk cloned from a ssd snapshot keeps the
// type of its source snapshot when disk_type is omitted, and that the type filled back by
// the remote api does not produce a diff which would force the disk to be rebuilt
func TestAccUCloudDisk_fromSnapshot(t *testing.T) {
	var diskSet udisk.UDiskDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk.bar",
		Providers:     testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			if err := testAccCheckDiskDestroy(s); err != nil {
				return err
			}
			return testAccCheckDiskSnapshotDestroy(s)
		},

		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfigFromSnapshot,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.bar", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.bar", "disk_type", "ssd_data_disk"),
					resource.TestCheckResourceAttr("ucloud_disk.bar", "disk_size", "20"),
					resource.TestCheckResourceAttrPair("ucloud_disk.bar", "snapshot_id", "ucloud_disk_snapshot.foo", "id"),
				),
			},

			// applying the same config again must be a no-op, otherwise the computed
			// disk_type is producing a permanent diff on a ForceNew attribute
			{
				Config: testAccDiskConfigFromSnapshot,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.bar", &diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.bar", "disk_type", "ssd_data_disk"),
				),
			},
		},
	})
}

// TestAccUCloudDisk_diskTypeCompatibility guards the change of disk_type from a defaulted
// attribute to a computed one: a config which omits disk_type, the shape used by the
// existing users, must stay diff free across applies
func TestAccUCloudDisk_diskTypeCompatibility(t *testing.T) {
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
				Config: testAccDiskConfigWithoutDiskType,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckDiskAttributes(&diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_type", "data_disk"),
				),
			},

			{
				Config: testAccDiskConfigWithoutDiskType,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_type", "data_disk"),
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

const testAccDiskConfigFromSnapshot = `
data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-from-snapshot-source"
	tag               = "tf-acc"
	disk_size         = 20
	disk_type         = "ssd_data_disk"
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "tf-acc-disk-from-snapshot"
}

# disk_type is intentionally omitted, it should follow the source snapshot
resource "ucloud_disk" "bar" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-from-snapshot-clone"
	tag               = "tf-acc"
	disk_size         = 20
	snapshot_id       = "${ucloud_disk_snapshot.foo.id}"
}
`

const testAccDiskConfigWithoutDiskType = `
data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-without-disk-type"
	tag               = "tf-acc"
	disk_size         = 10
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
