package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudSubnetsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSubnetsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_subnets.foo"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.0.name", "tf-acc-subnets-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.0.tag", "tf-acc"),
				),
			},
		},
	})
}

func TestAccUCloudSubnetsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSubnetsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_subnets.foo"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.0.name", "tf-acc-subnets-dataSource-ids"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.0.tag", "tf-acc"),
				),
			},
		},
	})
}

func TestAccUCloudSubnetsDataSource_VPCId(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSubnetsConfigVPCId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_subnets.foo"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.0.name", "tf-acc-subnets-dataSource-VPCId"),
					resource.TestCheckResourceAttr("data.ucloud_subnets.foo", "subnets.0.tag", "tf-acc"),
				),
			},
		},
	})
}

const testAccDataSubnetsConfig = `

variable "name" {
	default = "tf-acc-subnets-dataSource-basic"
}

resource "ucloud_vpc" "foo" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "${var.name}"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_subnets" "foo" {
	name_regex  = "${ucloud_subnet.foo.name}"
}
`

const testAccDataSubnetsConfigIds = `

variable "name" {
	default = "tf-acc-subnets-dataSource-ids"
}

resource "ucloud_vpc" "foo" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "${var.name}"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_subnets" "foo" {
	ids = ["${ucloud_subnet.foo.*.id}"]
}
`

const testAccDataSubnetsConfigVPCId = `

variable "name" {
	default = "tf-acc-subnets-dataSource-VPCId"
}

resource "ucloud_vpc" "foo" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "${var.name}"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_subnets" "foo" {
	name_regex  = "${ucloud_subnet.foo.name}"
	vpc_id = "${ucloud_vpc.foo.id}"
}
`
