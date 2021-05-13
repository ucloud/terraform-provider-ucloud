package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudUFSsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUFSsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_ufss.foo"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.0.name", "tf-acc-ufss-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.0.size", "500"),
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
				Config: testAccDataUFSsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_ufss.foo"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_ufss.foo", "ufss.0.size", "500"),
				),
			},
		},
	})
}

const testAccDataUFSsConfig = `

variable "name" {
	default = "tf-acc-ufss-dataSource-basic"
}

resource "ucloud_ufs" "foo" {
	name          = "${var.name}"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}

data "ucloud_ufss" "foo" {
	name_regex  = "${ucloud_ufs.foo.name}"
}
`

const testAccDataUFSsConfigIds = `

variable "name" {
	default = "tf-acc-ufss-dataSource-ids"
}

variable "instance_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

data "ucloud_zones" "default" {}

resource "ucloud_ufs" "foo" {
	name          = "${var.name}-${format(var.count_format, count.index+1)}"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
	count 		  = "${var.instance_count}"
}

data "ucloud_ufss" "foo" {
	ids = ucloud_ufs.foo.*.id
}
`
