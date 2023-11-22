package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/iam"
)

func TestAccUCloudIAMPolicyDocument_basic(t *testing.T) {
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
				Config: testAccDataIAMPolicyDocumentConfig,

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

const testAccDataIAMPolicyDocumentConfig = `
data "ucloud_iam_policy_document" foo {
  version = "1"
  statement {
    effect = "Allow"
    
    action = [
      "uhost:TerminateUHostInstance",
      "uhost:DeleteIsolationGroup",
    ]
    
    resource = [
      "ucs:uhost:*:123:instance/uhost-xxx",
    ]
  }
  statement {
    effect = "Allow"
    
    action = [
      "uhost:DescribeUHostInstance"
    ]
    
    resource = [
      "*",
    ]
  }
}
resource "ucloud_iam_policy" "foo" {
	name  = "tf-acc-iam-policy"
	comment = "comment"
    policy = data.ucloud_iam_policy_document.foo.json
	scope = "Project"
}
`
