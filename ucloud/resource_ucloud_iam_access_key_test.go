package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMAccessKey_update_status(t *testing.T) {
	var val iam.AccessKey

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
					testAccCheckIAMAccessKeyExists("ucloud_iam_access_key.foo", &val),
					testAccCheckIAMAccessKeyAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_access_key.foo", "user_name", "tf-acc-iam-user"),
				),
			},
			{
				Config: testAccIAMAccessKeyConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMAccessKeyExists("ucloud_iam_access_key.foo", &val),
					testAccCheckIAMAccessKeyAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_access_key.foo", "user_name", "tf-acc-iam-user"),
					resource.TestCheckResourceAttr("ucloud_iam_access_key.foo", "status", iamStatusInactive),
				),
			},
		},
	})
}

func testAccCheckIAMAccessKeyExists(n string, val *iam.AccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vpc id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeAccessKey(rs.Primary.Attributes["user_name"], rs.Primary.ID)

		log.Printf("[INFO] access key id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMAccessKeyAttributes(val *iam.AccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.AccessKeyID == "" {
			return fmt.Errorf("access key id is empty")
		}

		return nil
	}
}

func testAccCheckIAMAccessKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_iam_access_key" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeAccessKeyByID(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.AccessKeyID != "" {
			return fmt.Errorf("IAM access key still exist")
		}
	}

	return nil
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
