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

* `name` - (Required) The name of VPC.
* `cidr_blocks` - (Required) The CIDR blocks of VPC.
* `tag` - (Optional) A mapping of tags to assign to VPC, the default value is "Default"(means no tag assigned).
* `remark` - (Optional) The remarks of the VPC, the default value is "".

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation for VPC.
* `update_time` - The time whenever there is a change made to VPC.
* `network_info` - network_info is a nested type. network_info documented below.

The attribute (`network_info`) support the following:

* `cidr_block` - The CIDR block of the VPC.
