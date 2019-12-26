package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudImagesDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataImagesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_images.foo"),
				),
			},
		},
	})
}

func TestAccUCloudImagesDataSourceMostRecent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataImagesConfigMostRecent,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_images.foo"),
					resource.TestCheckResourceAttr("data.ucloud_images.foo", "images.#", "1"),
				),
			},
		},
	})
}

const testAccDataImagesConfig = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        = "base"
}
`

const testAccDataImagesConfigMostRecent = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        = "base"
	most_recent		  = true
}
`
