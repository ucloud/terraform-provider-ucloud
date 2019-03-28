package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudLBsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lbs.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.0.name", "tf-acc-lbs-dataSource-basic"),
				),
			},
		},
	})
}

func TestAccUCloudLBsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lbs.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.0.name", "tf-acc-lbs-dataSource-ids"),
				),
			},
		},
	})
}

func TestAccUCloudLBsDataSource_vpc(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBsConfigVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lbs.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.0.name", "tf-acc-lbs-dataSource-vpc"),
				),
			},
		},
	})
}

func TestAccUCloudLBsDataSource_subnet(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBsConfigVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lbs.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.0.name", "tf-acc-lbs-dataSource-subnet"),
				),
			},
		},
	})
}

const testAccDataLBsConfig = `
resource "ucloud_lb" "foo" {
	name    = "tf-acc-lbs-dataSource-basic"
	tag  	= "tf-acc"
}

data "ucloud_lbs" "foo" {
	name_regex  = "${ucloud_lb.foo.name}"
}
`
const testAccDataLBsConfigIds = `

resource "ucloud_lb" "foo" {
	name    = "tf-acc-lbs-dataSource-ids"
	tag  	= "tf-acc"
	count   = 2
}

data "ucloud_lbs" "foo" {
	ids = ["${ucloud_lb.foo.*.id}"]
}
`
const testAccDataLBsConfigVPC = `

variable "name" {
	default = "tf-acc-lbs-dataSource-vpc"
}

resource "ucloud_vpc" "default" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
  
resource "ucloud_subnet" "default" {
	name       = "${var.name}"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_lb" "foo" {
	name    = "${var.name}"
	tag  	= "tf-acc
	vpc_id  = "${ucloud_vpc.default.id}"
	subnet 	= "${ucloud_subnet.default.id}"
}

data "ucloud_lbs" "foo" {
	vpc_id  = "${ucloud_vpc.default.id}"
	name_regex  = "${ucloud_lb.foo.name}"
}
`
const testAccDataLBsConfigSubnet = `

variable "name" {
	default = "tf-acc-lbs-dataSource-subnet"
}

resource "ucloud_vpc" "default" {
	name        = "${var.name}"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
  
resource "ucloud_subnet" "default" {
	name       = "${var.name}"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_lb" "foo" {
	name    = "${var.name}"
	tag  	= "tf-acc
	vpc_id  = "${ucloud_vpc.default.id}"
	subnet 	= "${ucloud_subnet.default.id}"
}

data "ucloud_lbs" "foo" {
	subnet 	= "${ucloud_subnet.default.id}"
	name_regex  = "${ucloud_lb.foo.name}"
}
`
