package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMGroup_basic(t *testing.T) {
	var value iam.Group

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
					testAccCheckIAMGroupExists("ucloud_iam_group.foo", &value),
					testAccCheckIAMGroupAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "name", "tf-acc-iam-group"),
				),
			},
		},
	})
}

func TestAccUCloudIAMGroup_update_status(t *testing.T) {
	var value iam.Group

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
					testAccCheckIAMGroupExists("ucloud_iam_group.foo", &value),
					testAccCheckIAMGroupAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "name", "tf-acc-iam-group"),
				),
			},
			{
				Config: testAccIAMGroupConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupExists("ucloud_iam_group.foo", &value),
					testAccCheckIAMGroupAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "name", "tf-acc-iam-group"),
					resource.TestCheckResourceAttr("ucloud_iam_group.foo", "comment", "comment-update"),
				),
			},
		},
	})
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
