package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func testAccCheckIAMGroupMembershipExists(n string, user *[]iam.UserForGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeGroupMembership(rs.Primary.ID)

		log.Printf("[INFO] group name %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*user = ptr
		return nil
	}
}

func testAccCheckIAMGroupMembershipAttributes(val *[]iam.UserForGroup, size int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(*val) != size {
			return fmt.Errorf("length of val is not %v", size)
		}

		return nil
	}
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
