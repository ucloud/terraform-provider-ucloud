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
					resource.TestCheckResourceAttr("data.ucloud_lbs.foo", "lbs.#", "2"),
				),
			},
		},
	})
}

const testAccDataLBsConfig = `
resource "ucloud_lb" "foo" {
	name    = "tf-test-acc-lb"
	tag  	= "tf-acc"
	count   = 2
}

data "ucloud_lbs" "foo" {
	ids = ["${ucloud_lb.foo.*.id}"]
}
`
