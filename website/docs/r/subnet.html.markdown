---
layout: "ucloud"
page_title: "UCloud: ucloud_subnet"
sidebar_current: "docs-ucloud-resource-subnet"
description: |-
  Provides a Subnet resource under VPC resource.
---

# ucloud_subnet

Provides a Subnet resource under VPC resource.

## Example Usage

```hcl
resource "ucloud_vpc" "default" {
    name = "tf-example-vpc"
    tag  = "tf-example"

    # vpc network
    cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "example" {
    name = "tf-example-subnet"
    tag  = "tf-example"

    # subnet's network must be contained by vpc network
    # and a subnet must have least 8 ip addresses in it (netmask < 30).
    cidr_block = "192.168.1.0/24"
    vpc_id     = "${ucloud_vpc.default.id}"
}
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Required) The cidr block of the desired subnet, format in "0.0.0.0/0", such as: `192.168.0.0/24`.
* `vpc_id` - (Required) The id of the VPC that the desired subnet belongs to.
* `name` - (Optional) The name of the desired subnet. If not specified, terraform will autogenerate a name beginning with `tf-subnet`.
* `remark` - (Optional) The remarks of the subnet. (Default: `""`).
* `tag` - (Optional) A tag assigned to subnet, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of subnet, formatted in RFC3339 time string.
