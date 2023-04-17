---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_policy_document"
description: |-
  Generates an IAM policy document in JSON format for use with resources that expect policy documents such as ucloud_iam_policy.
---

# ucloud_iam_policy

Generates an IAM policy document in JSON format for use with resources that expect policy documents such as ucloud_iam_policy.

## Example Usage

```hcl
data "ucloud_iam_policy_document" foo {
  version = 1
  statement {
    effect = "Allow"
    
    action = [
      "uhost:TerminateUHostInstance",
      "uhost:DeleteIsolationGroup",
    ]
    
    resource = [
      "ucs:uhost:*:<company-id>:instance/uhost-xxx",
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
```

## Argument Reference

The following arguments are supported:

* `version` - (Optional) Version of the IAM policy document. Valid value is 1. Default value is 1.
* `statement` - (Optional) Statement of the IAM policy document. See the following Block statement.
* `output_file` - (Optional) File name where to save data source results (after running terraform plan).

#### Block statement

The statement supports the following:

* `effect` - (Optional) This parameter indicates whether the `action` is allowed. Valid values are `Allow` and `Deny`. Default value is `Allow`.
* `action` - (Required) Actions list of the IAM policy document. The format is `<product-name>:<api-name>`
* `resource` - (Optional) List of specific objects which will be authorized. Now UHost and UCDN resource are supported. The resource name can be `ucs:uhost:*:<company-id>:instance/<uhost-id>` or `ucs:ucdn:*:<company-id>:instance/<domain-id>`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `json` -  Policy JSON representation rendered based on the arguments above.
