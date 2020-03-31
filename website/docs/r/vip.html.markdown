---
layout: "ucloud"
page_title: "UCloud: ucloud_vip"
sidebar_current: "docs-ucloud-resource-vip"
description: |-
  Provides a VIP resource.
---

# ucloud_vip

Provides a VIP resource.

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
	vpc_id     = ucloud_vpc.foo.id
}
resource "ucloud_vip" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	subnet_id	= ucloud_subnet.foo.id
	name  	 	= "tf-acc-vip-basic"
	remark 		= "test"
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the VIP. 
* `subnet_id` - (Required, ForceNew) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. 

- - -

* `name` - (Optional) The name of VIP. If not specified, terraform will auto-generate a name beginning with `tf-vip-`.
* `tag` - (Optional) A tag assigned to VIP, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `remark` - (Optional) The remarks of the VIP. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ip_address` - The ip address of the VIP.
* `create_time` - The time of creation for VIP, formatted in RFC3339 time string.