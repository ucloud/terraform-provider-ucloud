package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudRepositoryImageTagsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryImageTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_repository_image_tags.foo"),
					resource.TestCheckResourceAttrSet("data.ucloud_repository_image_tags.foo", "repository_image_tags.#"),
					resource.TestCheckResourceAttr("data.ucloud_repository_image_tags.foo", "repository_image_tags.0.name", "latest"),
				),
			},
		},
	})
}

const testAccDataRepositoryImageTagsConfig = `
data "ucloud_repository_image_tags" "foo" {
	repository_name  = "ucloud"
	image_name = "nginx"
}
`
