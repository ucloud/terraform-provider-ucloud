---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_group"
description: |-
  Provides an IAM group resource.
---

# ucloud_iam_group

Provides an IAM group resource.

## Example Usage

```hcl
resource "ucloud_iam_group" "foo" {
	name  = "tf-acc-iam-group"
	comment = "comment"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, ForceNew) Name of the IAM group.
* `comment` - (Optional) Comment of the IAM group.

## Import
IAM group can be imported using group name, e.g.

```
$ terraform import ucloud_iam_group.example Administrator
```