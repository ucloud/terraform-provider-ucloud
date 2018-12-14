---
layout: "ucloud"
page_title: "UCloud: ucloud_vpc"
sidebar_current: "docs-ucloud-resource-vpc"
description: |-
  Provides a VPC resource.
---

# ucloud_vpc

Provides a VPC resource.

## Example Usage

```hcl
resource "ucloud_vpc" "example" {
    name = "tf-example-vpc"
    tag  = "tf-example"

    # vpc network
    cidr_blocks = ["192.168.0.0/16"]
}
```

## Argument Reference

The following arguments are supported:

* `cidr_blocks` - (Required) The CIDR blocks of VPC.
* `name` - (Optional) The name of VPC. If not specified, terraform will autogenerate a name beginning with `"tf-vpc"`.
* `tag` - (Optional) A mapping of tags to assign to VPC, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', and '.'. (Default: `"Default"`).
* `remark` - (Optional) The remarks of the VPC. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation for VPC, formatted in RFC3339 time string.
* `update_time` - The time whenever there is a change made to VPC, formatted in RFC3339 time string.
* `network_info` - It is a nested type which documented below.

The attribute (`network_info`) support the following:

* `cidr_block` - The CIDR block of the VPC.
