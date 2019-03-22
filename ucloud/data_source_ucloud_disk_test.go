package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
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
					resource.TestCheckResourceAttr("data.ucloud_disks.foo", "disks.#", "2"),
				),
			},
		},
	})
}

const testAccDataDisksConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "tf-acc-disks"
	tag               = "tf-acc"
	disk_size         = 10
	count 			  = 2
}

data "ucloud_disks" "foo" {
	ids = ["${ucloud_disk.foo.*.id}"]
}
`
