package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudInstancesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataInstancesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_instances.foo"),
					resource.TestCheckResourceAttr("data.ucloud_instances.foo", "instances.#", "2"),
				),
			},
		},
	})
}

const testAccDataInstancesConfig = `
resource "ucloud_instance" "foo" {
	count = 2

	availability_zone = "cn-sh2-02"
	image_id = "uimage-of3pac"
	root_password = "wA1234567"

	name = "testAccInstance"
	instance_type = "n-highcpu-1"
}

data "ucloud_instances" "foo" {
	ids = ["${ucloud_instance.foo.*.id}"]
}
`
