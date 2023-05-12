---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_group_membership"
description: |-
  Provides an IAM group membership resource.
---

# ucloud_iam_group_membership

Provides an IAM group membership resource.

## Example Usage

```hcl
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
}
resource "ucloud_iam_group_membership" "foo" {
	group_name = ucloud_iam_group.foo.name
	user_names = [
		ucloud_iam_user.foo.name
	]
}
```

## Argument Reference

The following arguments are supported:

* `group_name` - (Required, ForceNew) Name of the IAM group.
* `user_names` - (Required) Set of user name which will be added to group.

## Import
IAM group membership can be imported using group name, e.g.

```
$ terraform import ucloud_iam_group_membership.example Administrator
```