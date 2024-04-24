---
subcategory: "Label"
layout: "ucloud"
page_title: "UCloud: ucloud_label_resources"
description: |-
  Provides a list of resources with specific label.
---

# ucloud_labels

Provides a list of labels.

## Example Usage

```hcl
data "ucloud_label_resources" "foo" {
	key   = "key"
	value =  "value"
	resource_types = ["vip"]
	project_ids = ["org-xxx"]
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required, String) key of the label.s
* `value` - (Required, String) value of the label
* `resource_types` - (Required, String Array) types of the attached resources, for example uhost.
* `project_ids` - (Required, String Array) projects that own the attached resources, for example org-xxx.*
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `resources` - Resources with specific label and consists of following attribute
  * `id` - ID of the resource
  * `name` - Name of the resource
