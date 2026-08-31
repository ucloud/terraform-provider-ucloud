package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMProject_basic(t *testing.T) {
	var value iam.Project

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_project.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMProjectDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMProjectConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMProjectExists("ucloud_iam_project.foo", &value),
					testAccCheckIAMProjectAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_project.foo", "name", "tf-acc-iam-project"),
				),
			},
		},
	})
}

func TestAccUCloudIAMProject_update_status(t *testing.T) {
	var value iam.Project

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_project.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMProjectConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMProjectExists("ucloud_iam_project.foo", &value),
					testAccCheckIAMProjectAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_project.foo", "name", "tf-acc-iam-project"),
				),
			},
			{
				Config: testAccIAMProjectConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMProjectExists("ucloud_iam_project.foo", &value),
					testAccCheckIAMProjectAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_project.foo", "name", "tf-acc-iam-project-test"),
				),
			},
		},
	})
}

const testAccIAMProjectConfig = `
resource "ucloud_iam_project" "foo" {
	name  = "tf-acc-iam-project"
}
`

const testAccIAMProjectConfigUpdate = `
resource "ucloud_iam_project" "foo" {
	name  = "tf-acc-iam-project-test"
}
`
