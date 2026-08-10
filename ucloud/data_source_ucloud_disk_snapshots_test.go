package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudDiskSnapshotsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDiskSnapshotsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disk_snapshots.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.0.name", "tf-acc-disk-snapshots-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.0.disk_type", "data_disk"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.0.size", "20"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.0.status", "Normal"),
				),
			},
		},
	})
}

func TestAccUCloudDiskSnapshotsDataSource_diskId(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDiskSnapshotsConfigDiskId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disk_snapshots.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "total_count", "1"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.0.name", "tf-acc-disk-snapshots-dataSource-diskId"),
				),
			},
		},
	})
}

func TestAccUCloudDiskSnapshotsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDiskSnapshotsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disk_snapshots.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.#", "2"),
				),
			},
		},
	})
}

// TestAccUCloudDiskSnapshotsDataSource_withoutZone probes whether the remote api returns the
// snapshots of the whole region when the zone is omitted. The sdk marks the zone of
// DescribeUDiskSnapshot as optional while requiring it for the UDiskId filter, so the actual
// scope has to be settled by a real run before the availability_zone is documented as optional
func TestAccUCloudDiskSnapshotsDataSource_withoutZone(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDiskSnapshotsConfigWithoutZone,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disk_snapshots.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_disk_snapshots.foo", "snapshots.0.name", "tf-acc-disk-snapshots-dataSource-withoutZone"),
				),
			},
		},
	})
}

const testAccDataDiskSnapshotsConfig = `
variable "name" {
	default = "tf-acc-disk-snapshots-dataSource-basic"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 20
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "${var.name}"
}

data "ucloud_disk_snapshots" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "${ucloud_disk_snapshot.foo.name}"
}
`

const testAccDataDiskSnapshotsConfigDiskId = `
variable "name" {
	default = "tf-acc-disk-snapshots-dataSource-diskId"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 20
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "${var.name}"
}

data "ucloud_disk_snapshots" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk_snapshot.foo.disk_id}"
}
`

const testAccDataDiskSnapshotsConfigWithoutZone = `
variable "name" {
	default = "tf-acc-disk-snapshots-dataSource-withoutZone"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 20
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "${var.name}"
}

data "ucloud_disk_snapshots" "foo" {
	name_regex = "${ucloud_disk_snapshot.foo.name}"
}
`

const testAccDataDiskSnapshotsConfigIds = `
variable "name" {
	default = "tf-acc-disk-snapshots-dataSource-ids"
}

variable "snapshot_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 20
	snapshot_service  = true
}

resource "ucloud_disk_snapshot" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id           = "${ucloud_disk.foo.id}"
	name              = "${var.name}-${format(var.count_format, count.index+1)}"
	count             = "${var.snapshot_count}"
}

data "ucloud_disk_snapshots" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	ids               = ucloud_disk_snapshot.foo.*.id
}
`
