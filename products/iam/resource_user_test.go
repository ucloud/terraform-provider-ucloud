package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMUser_basic(t *testing.T) {
	var value iam.User

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_user.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMUserDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserExists("ucloud_iam_user.foo", &value),
					testAccCheckIAMUserAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "name", "tf-acc-iam-user"),
				),
			},
		},
	})
}

func TestAccUCloudIAMUser_update_status(t *testing.T) {
	var value iam.User

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_user.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMUserDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserExists("ucloud_iam_user.foo", &value),
					testAccCheckIAMUserAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "name", "tf-acc-iam-user"),
				),
			},
			{
				Config: testAccIAMUserConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserExists("ucloud_iam_user.foo", &value),
					testAccCheckIAMUserAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "name", "tf-acc-iam-user"),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "is_frozen", "true"),
				),
			},
		},
	})
}

const testAccIAMUserConfig = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
`

const testAccIAMUserConfigUpdate = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = true
}
`
