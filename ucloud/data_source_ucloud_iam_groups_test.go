package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudIAMGroupsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataIAMGroupsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_iam_groups.foo"),
					resource.TestCheckResourceAttr("data.ucloud_iam_groups.foo", "groups.#", "1"),
				),
			},
		},
	})
}

const testAccDataIAMGroupsConfig = `
data "ucloud_iam_groups" "foo" {
	name_regex        = "^Administrator$"
}
`
