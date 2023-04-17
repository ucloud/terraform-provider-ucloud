---
subcategory: "IAM"
layout: "ucloud"
page_title: "UCloud: ucloud_iam_projects"
description: |-
  Provides a list of IAM projects.
---

# ucloud_iam_groups

Provides a list of IAM projects.

## Example Usage

```hcl
data "ucloud_iam_projects" "foo" {
	name_regex = "^tf-acc-iam-project$"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) A regex string to filter resulting users by their names.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `projects` - A list of projects. Each element contains the following attributes
  * `name` - Name of the IAM project.
  * `id` - ID of the IAM project.
* `names` - A list of IAM project names.
