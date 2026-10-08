---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_route_table_rule"
description: |-
  Provides a Route Table Rule resource.
---

# ucloud_route_table_rule

Provides a Route Table Rule resource under a VPC route table. A route table rule describes how traffic destined for a
specific destination network segment is forwarded to a next hop (an instance or a private VIP).

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-example-subnet"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_vip" "foo" {
  vpc_id    = ucloud_vpc.foo.id
  subnet_id = ucloud_subnet.foo.id
  name      = "tf-example-vip"
  tag       = "tf-example"
}

resource "ucloud_route_table" "foo" {
  name   = "tf-example-route-table"
  tag    = "tf-example"
  vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table_rule" "foo" {
  route_table_id = ucloud_route_table.foo.id
  dst_addr       = "10.8.0.0/16"
  nexthop_type   = "vip"
  nexthop_id     = ucloud_vip.foo.id
  remark         = "tf-example-route-table-rule"
}
```

## Argument Reference

The following arguments are supported:

* `route_table_id` - (Required, ForceNew) The ID of the route table that the rule belongs to.
* `dst_addr` - (Required, ForceNew) The destination network segment of the route rule, an IPv4 address or CIDR network, e.g. `10.8.0.0/16` or `0.0.0.0/0`.
* `nexthop_type` - (Required, ForceNew) The type of the next hop, `instance` for a cloud host instance, `vip` for a private VIP.
* `nexthop_id` - (Required) The ID of the next hop resource. For `instance` it is a UHost instance ID, for `vip` it is a private VIP ID.

- - -

* `remark` - (Optional) The remarks of the route rule. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the route table rule.

## Import

Route Table Rule is not importable because its identity also depends on the parent route table ID.
