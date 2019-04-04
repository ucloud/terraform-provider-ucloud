package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudDisksDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDisksConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disks.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.name", "tf-acc-disks-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.disk_size", "10"),
				),
			},
		},
	})
}

func TestAccUCloudDisksDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDisksConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disks.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.name", "tf-acc-disks-dataSource-ids"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.disk_size", "10"),
				),
			},
		},
	})
}

func TestAccUCloudDisksDataSource_diskType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDisksConfigDiskType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disks.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.name", "tf-acc-disks-dataSource-diskType"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.disk_size", "10"),
				),
			},
		},
	})
}

const testAccDataDisksConfig = `

variable "name" {
	default = "tf-acc-disks-dataSource-basic"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 10
}

data "ucloud_disks" "foo" {
	name_regex  = "${ucloud_disk.foo.name}"
}
`

const testAccDataDisksConfigIds = `

variable "name" {
	default = "tf-acc-disks-dataSource-ids"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 10
	count 			  = 2
}

data "ucloud_disks" "foo" {
	ids = ["${ucloud_disk.foo.*.id}"]
}
`

const testAccDataDisksConfigDiskType = `

variable "name" {
	default = "tf-acc-disks-dataSource-diskType"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 10
}

data "ucloud_disks" "foo" {
	disk_type = "data_disk"
	name_regex  = "${ucloud_disk.foo.name}"
}
`
