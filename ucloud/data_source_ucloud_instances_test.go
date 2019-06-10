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
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        = "base"
}
  
resource "ucloud_instance" "foo" {
	count = 2

	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id      	  = "${data.ucloud_images.default.images.0.id}"
	root_password 	  = "wA1234567"
	name 		 	  = "tf-acc-instances"
	instance_type 	  = "n-highcpu-1"
}

data "ucloud_instances" "foo" {
	ids = ucloud_instance.foo.*.id
}
`
