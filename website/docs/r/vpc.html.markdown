---
layout: "ucloud"
page_title: "UCloud: ucloud_vpc"
sidebar_current: "docs-ucloud-resource-vpc"
description: |-
  Provides a VPC resource.
---

# ucloud_vpc

Provides a VPC resource.

~> **Note**  The network segment can only be created or deleted, can not perform both of them at the same time.
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

- - -

* `name` - (Optional, ForceNew) The name of VPC. If not specified, terraform will auto-generate a name beginning with `tf-vpc`.
* `tag` - (Optional, ForceNew) A tag assigned to VPC, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `remark` - (Optional, ForceNew) The remarks of the VPC. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource VPC.
* `create_time` - The time of creation for VPC, formatted in RFC3339 time string.
* `update_time` - The time whenever there is a change made to VPC, formatted in RFC3339 time string.
* `network_info` - It is a nested type which documented below.

- - -

The attribute (`network_info`) support the following:

* `cidr_block` - The CIDR block of the VPC.

## Import

VPC can be imported using the `id`, e.g.

```
$ terraform import ucloud_vpc.example uvnet-abc123456
```