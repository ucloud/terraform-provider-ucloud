package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMPolicy_basic(t *testing.T) {
	var value iam.IAMPolicy

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
					testAccCheckIAMPolicyExists("ucloud_iam_policy.foo", &value),
					testAccCheckIAMPolicyAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "name", "tf-acc-iam-policy"),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "comment", "comment"),
				),
			},
		},
	})
}

func TestAccUCloudIAMPolicy_update_status(t *testing.T) {
	var value iam.IAMPolicy

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
					testAccCheckIAMPolicyExists("ucloud_iam_policy.foo", &value),
					testAccCheckIAMPolicyAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "name", "tf-acc-iam-policy"),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "comment", "comment"),
				),
			},
			{
				Config: testAccIAMPolicyConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyExists("ucloud_iam_policy.foo", &value),
					testAccCheckIAMPolicyAttributes(&value),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "name", "tf-acc-iam-policy"),
					resource.TestCheckResourceAttr("ucloud_iam_policy.foo", "comment", "comment-update"),
				),
			},
		},
	})
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
