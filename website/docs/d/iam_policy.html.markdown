---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_policy"
description: |-
  Provides the detail of an existing IAM policy.
---

# ucloud_iam_policy

Provides the detail of an existing IAM policy.

## Example Usage

```hcl
data "ucloud_iam_policy" "foo" {
  name = "AdministratorAccess"
  type = "System"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the IAM policy.
* `type` - (Required) The type of the IAM policy, which can be `System` or `Custom`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `urn` - URN of the IAM policy.
* `comment` - Description of IAM policy
* `policy` - The policy document.
* `scope` - The policy scope, which value can be Project, Account or Mixed.
* `create_time` - The creation time of IAM policy, formatted in RFC3339 time string.