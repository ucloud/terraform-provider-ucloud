---
subcategory: "Label"
layout: "ucloud"
page_title: "UCloud: ucloud_label_attachment"
description: |-
  Provides a label attachment resource.
---

# ucloud_label_attachment

Provides a label attachment resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vip"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
	name       = "tf-acc-vip"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = "${ucloud_vpc.foo.id}"
}
resource "ucloud_vip" "foo" {
	vpc_id	 	= "${ucloud_vpc.foo.id}"
	subnet_id	= "${ucloud_subnet.foo.id}"
	name  	 	= "tf-acc-vip-basic-update"
	remark 		= "test-update"
	tag         = "tf-acc"
}
resource "ucloud_label" "foo" {
	key   = "tf-acc-label-key"
	value = "tf-acc-label-value"
}
resource "ucloud_label_attachment" "foo" {
	key   = "${ucloud_label.foo.key}"
	value = "${ucloud_label.foo.value}"
    resource = "${ucloud_vip.foo.id}"
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required, String, ForceNew) key of the label.
* `value` - (Required, String, ForceNew) value of the label
* `resource` - (Required, String, ForceNew) id of the attached resource, for example vip-xxx.