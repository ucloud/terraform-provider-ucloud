package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudEipsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataEipsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_eips.foo"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.name", "tf-acc-eips-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.share_bandwidth_package_id", ""),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.share_bandwidth_package_name", ""),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.share_bandwidth", "0"),
				),
			},
		},
	})
}

func TestAccUCloudEipsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataEipsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_eips.foo"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.bandwidth", "1"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.1.charge_type", "month"),
				),
			},
		},
	})
}

const testAccDataEipsConfig = `
variable "name" {
	default = "tf-acc-eips-dataSource-basic"
}

resource "ucloud_eip" "foo" {
	name          = "${var.name}"
	bandwidth     = 1
	internet_type = "bgp"
	duration      = 1
}

data "ucloud_eips" "foo" {
	name_regex  = "${ucloud_eip.foo.name}"
}
`

const testAccDataEipsConfigIds = `

variable "name" {
	default = "tf-acc-eips-dataSource-ids"
}

variable "instance_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

resource "ucloud_eip" "foo" {
	name          = "${var.name}-${format(var.count_format, count.index+1)}"
	bandwidth     = 1
	internet_type = "bgp"
	duration      = 1
	count 	      = var.instance_count
}

data "ucloud_eips" "foo" {
	ids = ucloud_eip.foo.*.id
}
`
