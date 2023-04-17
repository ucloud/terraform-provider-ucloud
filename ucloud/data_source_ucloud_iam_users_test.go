package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudIAMUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserConfig,
			},
			{
				Config: testAccDataIAMUsersConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_iam_users.foo"),
					resource.TestCheckResourceAttr("data.ucloud_iam_users.foo", "users.#", "1"),
				),
			},
		},
	})
}

const testAccDataIAMUsersConfig = `
data "ucloud_iam_users" "foo" {
	name_regex        = "^tf-acc-iam-user$"
}
`
