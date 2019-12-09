package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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

func TestAccUCloudDisksDataSource_rssd(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDisksConfigrssd,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_disks.foo"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.name", "tf-acc-disks-dataSource-rssd"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.disk_size", "10"),
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.0.disk_type", "rssd_data_disk"),
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
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex  = "${ucloud_disk.foo.name}"
}
`

const testAccDataDisksConfigIds = `

variable "name" {
	default = "tf-acc-disks-dataSource-ids"
}

variable "instance_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}-${format(var.count_format, count.index+1)}"
	tag               = "tf-acc"
	disk_size         = 10
	count 			  = "${var.instance_count}"
}

data "ucloud_disks" "foo" {
	ids = ucloud_disk.foo.*.id
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

const testAccDataDisksConfigrssd = `

variable "name" {
	default = "tf-acc-disks-dataSource-rssd"
}

resource "ucloud_disk" "foo" {
	availability_zone = "cn-bj2-05"
	name              = "${var.name}"
	tag               = "tf-acc"
	disk_size         = 10
	disk_type         = "rssd_data_disk"
}

data "ucloud_disks" "foo" {
	availability_zone = "cn-bj2-05"
	disk_type = "rssd_data_disk"
	name_regex  = "${ucloud_disk.foo.name}"
}
`
