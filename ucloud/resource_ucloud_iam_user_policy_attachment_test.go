package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMUserPolicyAttachment_basic(t *testing.T) {
	var val iam.Policy

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_user_policy_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMUserPolicyAttachmentDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicyAttachmentExists("ucloud_iam_user_policy_attachment.foo", &val),
					testAccCheckIAMUserPolicyAttachmentAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_user_policy_attachment.foo", "user_name", "tf-acc-iam-user"),
				),
			},
		},
	})
}

func testAccCheckIAMUserPolicyAttachmentExists(n string, val *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user name is empty")
		}
		userName, policyURN, projectID, err := extractUCloudIAMUserPolicyAttachmentID(rs.Primary.ID)

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeIAMUserPolicyAttachment(userName, policyURN, projectID)

		log.Printf("[INFO] user name id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMUserPolicyAttachmentAttributes(val *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.PolicyName == "" {
			return fmt.Errorf("policy name is empty")
		}

		return nil
	}
}

func testAccCheckIAMUserPolicyAttachmentDestroy(s *terraform.State) error {
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

const testAccIAMUserPolicyAttachmentConfig = `
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_project" "foo" {
	name  = "tf-acc-iam-project"
}
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
	scope_type = "Project"
}
resource "ucloud_iam_user_policy_attachment" "foo" {
	user_name  = ucloud_iam_user.foo.name
	policy_urn = ucloud_iam_policy.foo.urn
	project_id = ucloud_iam_project.foo.id
}
`
