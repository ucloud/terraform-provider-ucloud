---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_route_table_association"
description: |-
  Provides a Route Table Association resource.
---

# ucloud_route_table_association

Provides a resource to bind a subnet to a route table. A subnet is always bound to a route table; when it is not
bound to any custom route table it stays on its VPC default route table. Deleting this resource rebinds the subnet
back to the default route table.

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

resource "ucloud_route_table" "foo" {
  name   = "tf-example-route-table"
  tag    = "tf-example"
  vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table_association" "foo" {
  subnet_id      = ucloud_subnet.foo.id
  route_table_id = ucloud_route_table.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required, ForceNew) The ID of the subnet to bind. It is also used as the resource ID.
* `route_table_id` - (Required) The ID of the route table to bind the subnet to. Changing this moves the subnet to the new route table.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the subnet that is bound.

## Import

Route Table Association can be imported using the `subnet_id`, e.g.

```
terraform import ucloud_route_table_association.example subnet-abc123456
```
