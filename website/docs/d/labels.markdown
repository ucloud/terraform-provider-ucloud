---
subcategory: "Label"
layout: "ucloud"
page_title: "UCloud: ucloud_labels"
description: |-
  Provides a list of labels.
---

# ucloud_labels

Provides a list of labels.

## Example Usage

```hcl
data "ucloud_labels" "foo" {
	key_regex   = "^key$"
}
```

## Argument Reference

The following arguments are supported:

*  `key_regex` - (Optional) A regex string to filter the returned users by their keys.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `key` - Key of the label
* `value` - Value of the label
* `projects` - Projects which have attached resources and consists of following attribute
  * `id` - ID of the project
  * `name` - Name of the project
  * `resource_types` - Array of strings, which are resource types with query permission for current account
  * `disabled_resource_types` - Array of strings, which are resource types without query permission for current account
