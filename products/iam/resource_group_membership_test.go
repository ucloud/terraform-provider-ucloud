package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMGroupMembership_basic(t *testing.T) {
	var users []iam.UserForGroup

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_group_membership.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupMembershipConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupMembershipExists("ucloud_iam_group_membership.foo", &users),
					testAccCheckIAMGroupMembershipAttributes(&users, 1),
					resource.TestCheckResourceAttr("ucloud_iam_group_membership.foo", "group_name", "tf-acc-iam-group"),
				),
			},
		},
	})
}

func TestAccUCloudIAMGroupMembership_update_status(t *testing.T) {
	var users []iam.UserForGroup

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_group_membership.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupMembershipConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupMembershipExists("ucloud_iam_group_membership.foo", &users),
					testAccCheckIAMGroupMembershipAttributes(&users, 1),
					resource.TestCheckResourceAttr("ucloud_iam_group_membership.foo", "group_name", "tf-acc-iam-group"),
					resource.TestCheckResourceAttr("ucloud_iam_group_membership.foo", "user_names.#", "1")),
			},
			{
				Config: testAccIAMGroupMembershipConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupMembershipExists("ucloud_iam_group_membership.foo", &users),
					testAccCheckIAMGroupMembershipAttributes(&users, 0),
					resource.TestCheckResourceAttr("ucloud_iam_group_membership.foo", "group_name", "tf-acc-iam-group"),
					resource.TestCheckResourceAttr("ucloud_iam_group_membership.foo", "user_names.#", "0")),
			},
		},
	})
}

const testAccIAMGroupMembershipConfig = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
}
resource "ucloud_iam_group_membership" "foo" {
	group_name = ucloud_iam_group.foo.name
	user_names = [
		ucloud_iam_user.foo.name
	]
}
`

const testAccIAMGroupMembershipConfigUpdate = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
}
resource "ucloud_iam_group_membership" "foo" {
	group_name = ucloud_iam_group.foo.name
	user_names = []
}
`
