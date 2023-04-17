package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMProject_basic(t *testing.T) {
	var val iam.Project

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
					testAccCheckIAMProjectExists("ucloud_iam_project.foo", &val),
					testAccCheckIAMProjectAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_project.foo", "name", "tf-acc-iam-project"),
				),
			},
		},
	})
}

func TestAccUCloudIAMProject_update_status(t *testing.T) {
	var val iam.Project

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
					testAccCheckIAMProjectExists("ucloud_iam_project.foo", &val),
					testAccCheckIAMProjectAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_project.foo", "name", "tf-acc-iam-project"),
				),
			},
			{
				Config: testAccIAMProjectConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMProjectExists("ucloud_iam_project.foo", &val),
					testAccCheckIAMProjectAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_project.foo", "name", "tf-acc-iam-project-test"),
				),
			},
		},
	})
}

func testAccCheckIAMProjectExists(n string, val *iam.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("project id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeIAMProjectById(rs.Primary.ID)

		log.Printf("[INFO] project id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMProjectAttributes(val *iam.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.ProjectName == "" {
			return fmt.Errorf("project name is empty")
		}

		return nil
	}
}

func testAccCheckIAMProjectDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_iam_project" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeIAMProjectById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.ProjectName != "" {
			return fmt.Errorf("project still exist")
		}
	}

	return nil
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
