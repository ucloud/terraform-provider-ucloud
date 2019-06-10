package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudLBSSLsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBSSLsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lb_ssls.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lb_ssls.foo", "lb_ssls.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_lb_ssls.foo", "lb_ssls.0.name", "tf-acc-lb-ssls-dataSource-basic"),
				),
			},
		},
	})
}

func TestAccUCloudLBSSLsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataLBSSLsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_lb_ssls.foo"),
					resource.TestCheckResourceAttr("data.ucloud_lb_ssls.foo", "lb_ssls.#", "2"),
				),
			},
		},
	})
}

const testAccDataLBSSLsConfig = `

variable "name" {
	default = "tf-acc-lb-ssls-dataSource-basic"
}

resource "ucloud_lb_ssl" "foo" {
	name 		= "${var.name}"
	private_key = "${file("test-fixtures/private.key")}"
	user_cert 	= "${file("test-fixtures/user.crt")}"
	ca_cert 	= "${file("test-fixtures/ca.crt")}"
}

data "ucloud_lb_ssls" "foo" {
	name_regex  = "${ucloud_lb_ssl.foo.name}"
}
`

const testAccDataLBSSLsConfigIds = `

variable "name" {
	default = "tf-acc-lb-ssls-dataSource-ids"
}

variable "instance_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}


resource "ucloud_lb_ssl" "foo" {
	name        = "${var.name}-${format(var.count_format, count.index+1)}"
	private_key = "${file("test-fixtures/private.key")}"
	user_cert 	= "${file("test-fixtures/user.crt")}"
	ca_cert 	= "${file("test-fixtures/ca.crt")}"
	count 		= "${var.instance_count}"
}

data "ucloud_lb_ssls" "foo" {
	ids = ucloud_lb_ssl.foo.*.id
}
`
