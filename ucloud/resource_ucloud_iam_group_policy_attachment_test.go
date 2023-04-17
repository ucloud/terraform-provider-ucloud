package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMGroupPolicyAttachment_basic(t *testing.T) {
	var val iam.Policy

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
					testAccCheckIAMGroupPolicyAttachmentExists("ucloud_iam_group_policy_attachment.foo", &val),
					testAccCheckIAMGroupPolicyAttachmentAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_iam_group_policy_attachment.foo", "group_name", "tf-acc-iam-group"),
				),
			},
		},
	})
}

func testAccCheckIAMGroupPolicyAttachmentExists(n string, val *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("group policy attachment is empty")
		}
		groupName, policyURN, projectID, err := extractUCloudIAMGroupPolicyAttachmentID(rs.Primary.ID)

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeIAMGroupPolicyAttachment(groupName, policyURN, projectID)

		log.Printf("[INFO] group policy attachment id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckIAMGroupPolicyAttachmentAttributes(val *iam.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.PolicyName == "" {
			return fmt.Errorf("policy name is empty")
		}

		return nil
	}
}

func testAccCheckIAMGroupPolicyAttachmentDestroy(s *terraform.State) error {
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
