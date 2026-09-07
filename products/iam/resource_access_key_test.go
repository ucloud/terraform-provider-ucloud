package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMAccessKey_update_status(t *testing.T) {
	var value iam.AccessKey

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_access_key.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMUserDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMAccessKeyConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMAccessKeyExists("ucloud_iam_access_key.foo", &value),
					testAccCheckIAMAccessKeyAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_access_key.foo", "user_name", "tf-acc-iam-user"),
				),
			},
			{
				Config: testAccIAMAccessKeyConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMAccessKeyExists("ucloud_iam_access_key.foo", &value),
					testAccCheckIAMAccessKeyAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_access_key.foo", "user_name", "tf-acc-iam-user"),
					resource.TestCheckResourceAttr("ucloud_iam_access_key.foo", "status", iamStatusInactive),
				),
			},
		},
	})
}

const testAccIAMAccessKeyConfig = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_access_key" "foo" {
	user_name  = "${ucloud_iam_user.foo.name}"
}
`

const testAccIAMAccessKeyConfigUpdate = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_access_key" "foo" {
	user_name  = "${ucloud_iam_user.foo.name}"
	status = "Inactive"
}
`
