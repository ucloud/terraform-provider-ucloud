package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMGroupPolicyAttachment_basic(t *testing.T) {
	var value iam.Policy

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_iam_group_policy_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupPolicyAttachmentDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyAttachmentExists("ucloud_iam_group_policy_attachment.foo", &value),
					testAccCheckIAMGroupPolicyAttachmentAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_group_policy_attachment.foo", "group_name", "tf-acc-iam-group"),
				),
			},
		},
	})
}

const testAccIAMGroupPolicyAttachmentConfig = `
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
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
resource "ucloud_iam_group_policy_attachment" "foo" {
	group_name  = ucloud_iam_group.foo.name
	policy_urn = ucloud_iam_policy.foo.urn
	project_id = ucloud_iam_project.foo.id
}
`
