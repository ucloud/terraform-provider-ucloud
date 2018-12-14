package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
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
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.#", "2"),
				),
			},
		},
	})
}

const testAccDataEipsConfig = `
resource "ucloud_eip" "foo" {
	count         = 2
	name          = "tf-test-acc-eip"
	bandwidth     = 1
	internet_type = "bgp"
	duration      = 1
}

data "ucloud_eips" "foo" {
	ids = ["${ucloud_eip.foo.*.id}"]
}
`
