package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
)

func TestAccUCloudDiskSnapshot_basic(t *testing.T) {
	var snapshotSet udisk.UDiskSnapshotSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk_snapshot.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskSnapshotDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDiskSnapshotConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskSnapshotExists("ucloud_disk_snapshot.foo", &snapshotSet),
					testAccCheckDiskSnapshotAttributes(&snapshotSet),
					resource.TestCheckResourceAttr("ucloud_disk_snapshot.foo", "name", "tf-acc-disk-snapshot-basic"),
					resource.TestCheckResourceAttr("ucloud_disk_snapshot.foo", "size", "20"),
					resource.TestCheckResourceAttr("ucloud_disk_snapshot.foo", "disk_type", "data_disk"),
					resource.TestCheckResourceAttr("ucloud_disk_snapshot.foo", "status", "Normal"),
				),
			},
		},
	})
}

func TestAccUCloudDiskSnapshot_ssd(t *testing.T) {
	var snapshotSet udisk.UDiskSnapshotSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk_snapshot.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskSnapshotDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDiskSnapshotConfigSSD,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskSnapshotExists("ucloud_disk_snapshot.foo", &snapshotSet),
					testAccCheckDiskSnapshotAttributes(&snapshotSet),
					resource.TestCheckResourceAttr("ucloud_disk_snapshot.foo", "disk_type", "ssd_data_disk"),
					resource.TestCheckResourceAttr("ucloud_disk_snapshot.foo", "status", "Normal"),
				),
			},
		},
	})
}

func testAccCheckDiskSnapshotExists(n string, snapshotSet *udisk.UDiskSnapshotSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("disk snapshot id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeSnapshotById(rs.Primary.ID)

		log.Printf("[INFO] disk snapshot id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*snapshotSet = *ptr
		return nil
	}
}

func testAccCheckDiskSnapshotAttributes(snapshotSet *udisk.UDiskSnapshotSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if snapshotSet.SnapshotId == "" {
			return fmt.Errorf("disk snapshot id is empty")
		}

		if snapshotSet.Status != snapshotStatusNormal {
			return fmt.Errorf("disk snapshot %q is expected to be %q, got %q", snapshotSet.SnapshotId, snapshotStatusNormal, snapshotSet.Status)
		}

		return nil
	}
}

func testAccCheckDiskSnapshotDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_disk_snapshot" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeSnapshotById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.SnapshotId != "" {
			return fmt.Errorf("disk snapshot still exist")
		}
	}

	return nil
}

const testAccDiskSnapshotConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-snapshot-basic"
	tag               = "tf-acc"
	disk_size         = 20
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "tf-acc-disk-snapshot-basic"
}
`

const testAccDiskSnapshotConfigSSD = `
data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disk-snapshot-ssd"
	tag               = "tf-acc"
	disk_size         = 20
	disk_type         = "ssd_data_disk"
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "tf-acc-disk-snapshot-ssd"
}
`
