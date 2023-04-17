---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_users"
description: |-
  Provides a list of IAM users.
---

# ucloud_iam_groups

Provides a list of IAM users.

## Example Usage

```hcl
data "ucloud_iam_users" "foo" {
  name_regex = "^tf-acc-iam-user$"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) A regex string to filter the returned users by their names.
* `group_name` - (Optional) Filter results by a specific group name. Returned users are in the specified group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `users` - A list of users. Each element contains the following attributes
  * `name` - Name of the IAM user.
  * `display_name` - Name of the IAM user which for display.
  * `email` - Email of the IAM user.
  * `status` - Status of the IAM user Active or Inactive.
  * `login_enable` - true or false.
* `names` - A list of IAM user names.
