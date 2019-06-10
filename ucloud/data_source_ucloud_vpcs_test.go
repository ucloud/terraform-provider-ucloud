package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudVPCsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataVPCsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_vpcs.foo"),
					resource.TestCheckResourceAttr("data.ucloud_vpcs.foo", "vpcs.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_vpcs.foo", "vpcs.0.name", "tf-acc-vpcs-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_vpcs.foo", "vpcs.0.tag", "tf-acc"),
				),
			},
		},
	})
}

func TestAccUCloudVPCsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataVPCsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_vpcs.foo"),
					resource.TestCheckResourceAttr("data.ucloud_vpcs.foo", "vpcs.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_vpcs.foo", "vpcs.0.name", "tf-acc-vpcs-dataSource-ids"),
					resource.TestCheckResourceAttr("data.ucloud_vpcs.foo", "vpcs.0.tag", "tf-acc"),
				),
			},
		},
	})
}

const testAccDataVPCsConfig = `

variable "name" {
	default = "tf-acc-vpcs-dataSource-basic"
}

resource "ucloud_vpc" "foo" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

data "ucloud_vpcs" "foo" {
	name_regex  = "${ucloud_vpc.foo.name}"
}
`

const testAccDataVPCsConfigIds = `

variable "name" {
	default = "tf-acc-vpcs-dataSource-ids"
}

resource "ucloud_vpc" "foo" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

data "ucloud_vpcs" "foo" {
	ids = ucloud_vpc.foo.*.id
}
`
