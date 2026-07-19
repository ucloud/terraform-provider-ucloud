package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudUK8SImagesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataUK8SImagesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_uk8s_images.foo"),
				),
			},
		},
	})
}

const testAccDataUK8SImagesConfig = `
data "ucloud_zones" "default" {
}

data "ucloud_uk8s_images" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	host_type         = "uhost"
}
`
