---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_policy"
description: |-
  Provides an IAM custom policy resource.
---

# ucloud_iam_policy

Provides an IAM custom policy resource.

## Example Usage

```hcl
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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, ForceNew) The name of the policy.
* `comment` - (Optional) Comments of the IAM policy.
* `policy` - (Required) The policy document. This is a JSON formatted string.
* `scope` - (Optional, ForceNew) The policy scope, which value can be Project or Account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of policy, formatted in RFC3339 time string.
* `urn` - URN of the policy.

## Import
IAM group can be imported using policy name, e.g.

```
$ terraform import ucloud_iam_policy.example uhost-policy
```