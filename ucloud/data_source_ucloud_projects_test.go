package ucloud

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudProjectsDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataProjectsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_projects.foo"),
					resource.TestMatchResourceAttr("data.ucloud_projects.foo", "projects.0.name", regexp.MustCompile(`^.{1,}$`)),
				),
			},
		},
	})
}

const testAccDataProjectsConfig = `
data "ucloud_projects" "foo" {
}
`
