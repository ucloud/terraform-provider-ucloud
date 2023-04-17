---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_groups"
description: |-
  Provides a list of IAM groups.
---

# ucloud_iam_groups

Provides a list of IAM groups.

## Example Usage

```hcl
data "ucloud_iam_groups" "foo" {
  name_regex  = "^Administrator$"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) A regex string to filter the returned groups by their names.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `groups` - A list of groups. Each element contains the following attributes
  * `name` - Name of the group.
  * `comment` - Comment of the group.
* `names` - A list of IAM user group names.
