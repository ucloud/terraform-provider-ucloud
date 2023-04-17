---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_user"
description: |-
  Provides an IAM user resource.
---

# ucloud_iam_user

Provides an IAM user resource.

## Example Usage

```hcl
resource "ucloud_iam_user" "foo" {
	name  = "tf-acc-iam-user"
	login_enable = false
	is_frozen = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, ForceNew) Name of the IAM user.
* `display_name` - (Optional) Name of the IAM user which for display.
* `email` - (Optional) Email of the IAM user.
* `is_frozen` - (Optional) true or false, default is false.
* `login_enable` - (Optional) true or false, default is true.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `status` - Active, Inactive or Frozen.

## Import
IAM group can be imported using name of user, e.g.

```
$ terraform import ucloud_iam_user.example Administrator
```