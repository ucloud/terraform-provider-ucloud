package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudIAMPolicyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataIAMPolicyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_iam_policy.foo"),
					resource.TestCheckResourceAttr("data.ucloud_iam_policy.foo", "urn", "ucs:iam::ucs:policy/AdministratorAccess"),
				),
			},
		},
	})
}

const testAccDataIAMPolicyConfig = `
data "ucloud_iam_policy" "foo" {
	name        = "AdministratorAccess"
	type		= "System"
}
`
