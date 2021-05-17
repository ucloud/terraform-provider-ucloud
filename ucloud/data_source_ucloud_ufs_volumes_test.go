package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudUFSVolumesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUFSVolumesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_ufs_volumes.foo"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.0.name", "tf-acc-ufss-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.0.size", "500"),
				),
			},
		},
	})
}

func TestAccUCloudUFSsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUFSVolumesConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_ufs_volumes.foo"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_ufs_volumes.foo", "ufs_volumes.0.size", "500"),
				),
			},
		},
	})
}

const testAccDataUFSVolumesConfig = `

variable "name" {
	default = "tf-acc-ufs-volumes-dataSource-basic"
}

resource "ucloud_ufs_volume" "foo" {
	name          = "${var.name}"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}

data "ucloud_ufs_volumes" "foo" {
	name_regex  = "${ucloud_ufs.foo.name}"
}
`

const testAccDataUFSVolumesConfigIds = `

variable "name" {
	default = "tf-acc-ufs-volumes-dataSource-ids"
}

variable "instance_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

data "ucloud_zones" "default" {}

resource "ucloud_ufs_volume" "foo" {
	name          = "${var.name}-${format(var.count_format, count.index+1)}"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
	count 		  = "${var.instance_count}"
}

data "ucloud_ufs_volumes" "foo" {
	ids = ucloud_ufs.foo.*.id
}
`
