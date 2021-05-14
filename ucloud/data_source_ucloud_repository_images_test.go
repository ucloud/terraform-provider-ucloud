package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudRepositoryImagesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryImagesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_repository_images.foo"),
					resource.TestCheckResourceAttr("data.ucloud_repository_images.foo", "repository_images.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_repository_images.foo", "repository_images.0.name", "nginx"),
				),
			},
		},
	})
}

const testAccDataRepositoryImagesConfig = `
data "ucloud_repository_images" "foo" {
	name_regex  = "nginx"
	repository_name   = "ucloud"
}
`
