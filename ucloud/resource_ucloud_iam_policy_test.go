package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMPolicy_basic(t *testing.T) {
	var val iam.IAMPolicy

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_policy.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyExists("ucloud_iam_policy.foo", &val),
					testAccCheckIAMPolicyAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "name", "tf-acc-iam-policy"),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "comment", "comment"),
				),
			},
		},
	})
}

func TestAccUCloudIAMPolicy_update_status(t *testing.T) {
	var val iam.IAMPolicy

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_policy.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyExists("ucloud_iam_policy.foo", &val),
					testAccCheckIAMPolicyAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "name", "tf-acc-iam-policy"),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "comment", "comment"),
				),
			},
			{
				Config: testAccIAMPolicyConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyExists("ucloud_iam_policy.foo", &val),
					testAccCheckIAMPolicyAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "name", "tf-acc-iam-policy"),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "comment", "comment-update"),
				),
			},
		},
	})
}

func testAccCheckIAMPolicyExists(n string, val *iam.IAMPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("policy name is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeIAMPolicyByName(rs.Primary.ID, "User")

		log.Printf("[INFO] policy name %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMPolicyAttributes(val *iam.IAMPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.PolicyName == "" {
			return fmt.Errorf("policy name is empty")
		}

		return nil
	}
}

func testAccCheckIAMPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_iam_policy" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeIAMPolicyByName(rs.Primary.ID, "User")

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.PolicyName != "" {
			return fmt.Errorf("policy still exist")
		}
	}

	return nil
}

const testAccIAMPolicyConfig = `
resource "ucloud_iam_policy" "foo" {
	name  = "tf-acc-iam-policy"
	comment = "comment"
    policy = jsonencode({
      Version = "1"
      Statement = [
      {
        Action = [
          "*",
        ]
        Effect   = "Allow"
        Resource = ["*"]
      },
      ]
    })
	scope = "Project"
}
`

const testAccIAMPolicyConfigUpdate = `
resource "ucloud_iam_policy" "foo" {
	name  = "tf-acc-iam-policy"
	comment = "comment-update"
    policy = jsonencode({
      Version = "1"
      Statement = [
      {
        Action = [
          "*",
        ]
        Effect   = "Allow"
        Resource = ["*"]
      },
      ]
    })
	scope = "Project"
}
`
