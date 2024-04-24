---
subcategory: "Label"
layout: "ucloud"
page_title: "UCloud: ucloud_label"
description: |-
  Provides a label resource.
---

# ucloud_label

Provides a label resource.

## Example Usage

```hcl
resource "ucloud_label" "foo" {
	key  = "tf-acc-label-key"
	value  = "tf-acc-label-value"
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required, String, ForceNew) key of the label.
* `value` - (Required, String, ForceNew) value of the label
