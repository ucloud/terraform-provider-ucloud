---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_project"
description: |-
  Provides an IAM project resource.
---

# ucloud_iam_project

Provides an IAM project resource.

## Example Usage

```hcl
resource "ucloud_iam_project" "foo" {
	name  = "tf-acc-iam-project"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the IAM project.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of project, formatted in RFC3339 time string.

## Import
IAM group can be imported using project ID, e.g.

```
$ terraform import ucloud_iam_project.example org-xxx
```