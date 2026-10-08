---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_route_table"
description: |-
  Provides a Route Table resource.
---

# ucloud_route_table

Provides a Route Table resource under a VPC.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_route_table" "foo" {
  name   = "tf-example-route-table"
  tag    = "tf-example"
  remark = "tf-example-route-table"
  vpc_id = ucloud_vpc.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, ForceNew) The ID of the VPC that the route table belongs to.

- - -

* `name` - (Optional) The name of the route table, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_' and '.'. If not specified, terraform will auto-generate a name beginning with `tf-route-table-`.
* `tag` - (Optional) A tag assigned to the route table, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or an empty string is filled in, then the default tag will be assigned. (Default: `Default`).
* `remark` - (Optional) The remarks of the route table. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the route table.
* `vpc_name` - The name of the VPC that the route table belongs to.
* `route_table_type` - The type of the route table. `1` is the default route table, `0` is a custom route table.
* `subnet_count` - The number of subnets bound to the route table.
* `subnet_ids` - The list of subnet IDs bound to the route table.
* `create_time` - The time of creation of the route table, formatted in RFC3339 time string.

## Import

Route Table can be imported using the `id`, e.g.

```
$ terraform import ucloud_route_table.example routetable-abc123456
```
