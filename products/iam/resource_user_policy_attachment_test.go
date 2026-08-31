package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMUserPolicyAttachment_basic(t *testing.T) {
	var value iam.Policy

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
					testAccCheckIAMUserPolicyAttachmentExists("ucloud_iam_user_policy_attachment.foo", &value),
					testAccCheckIAMUserPolicyAttachmentAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_user_policy_attachment.foo", "user_name", "tf-acc-iam-user"),
				),
			},
		},
	})
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
