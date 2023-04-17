package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMUser_basic(t *testing.T) {
	var val iam.User

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
					testAccCheckIAMUserExists("ucloud_iam_user.foo", &val),
					testAccCheckIAMUserAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "name", "tf-acc-iam-user"),
				),
			},
		},
	})
}

func TestAccUCloudIAMUser_update_status(t *testing.T) {
	var val iam.User

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
					testAccCheckIAMUserExists("ucloud_iam_user.foo", &val),
					testAccCheckIAMUserAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "name", "tf-acc-iam-user"),
				),
			},
			{
				Config: testAccIAMUserConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserExists("ucloud_iam_user.foo", &val),
					testAccCheckIAMUserAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "name", "tf-acc-iam-user"),
					resource.TestCheckResourceAttr("ucloud_iam_user.foo", "is_frozen", "true"),
				),
			},
		},
	})
}

func testAccCheckIAMUserExists(n string, val *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user name is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUser(rs.Primary.ID)

		log.Printf("[INFO] user name id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMUserAttributes(val *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.UserName == "" {
			return fmt.Errorf("user name is empty")
		}

		return nil
	}
}

func testAccCheckIAMUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_iam_user" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUser(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.UserName != "" {
			return fmt.Errorf("user still exist")
		}
	}

	return nil
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
