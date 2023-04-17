package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudIAMProjectsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMProjectConfig,
			},
			{
				Config: testAccDataIAMProjectsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_iam_projects.foo"),
					resource.TestCheckResourceAttr("data.ucloud_iam_projects.foo", "projects.#", "1"),
				),
			},
		},
	})
}

const testAccDataIAMProjectsConfig = `
data "ucloud_iam_projects" "foo" {
	name_regex        = "^tf-acc-iam-project$"
}
`
