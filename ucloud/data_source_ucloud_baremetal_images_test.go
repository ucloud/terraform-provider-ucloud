package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudBareMetalImagesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataBareMetalImagesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_baremetal_images.foo"),
				),
			},
		},
	})
}

const testAccDataBareMetalImagesConfig = `
data "ucloud_zones" "default" {
}

data "ucloud_baremetal_images" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^Ubuntu 16.04"
	image_type        = "base"
    os_type           = "Ubuntu"
    ids               = ["pimg-cs-xev4rz"]
}
output "image_id" {
  value = data.ucloud_baremetal_images.foo.images[0].id
}
`
