package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMGroup_basic(t *testing.T) {
	var val iam.Group

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupExists("ucloud_iam_group.foo", &val),
					testAccCheckIAMGroupAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "name", "tf-acc-iam-group"),
				),
			},
		},
	})
}

func TestAccUCloudIAMGroup_update_status(t *testing.T) {
	var val iam.Group

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupExists("ucloud_iam_group.foo", &val),
					testAccCheckIAMGroupAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "name", "tf-acc-iam-group"),
				),
			},
			{
				Config: testAccIAMGroupConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupExists("ucloud_iam_group.foo", &val),
					testAccCheckIAMGroupAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "name", "tf-acc-iam-group"),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "comment", "comment-update"),
				),
			},
		},
	})
}

func testAccCheckIAMGroupExists(n string, val *iam.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("group name is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeGroup(rs.Primary.ID)

		log.Printf("[INFO] group name %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMGroupAttributes(val *iam.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.GroupName == "" {
			return fmt.Errorf("group name is empty")
		}

		return nil
	}
}

func testAccCheckIAMGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_iam_group" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeGroup(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.GroupName != "" {
			return fmt.Errorf("group still exist")
		}
	}

	return nil
}

const testAccIAMGroupConfig = `
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
}
`

const testAccIAMGroupConfigUpdate = `
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment-update"
}
`
