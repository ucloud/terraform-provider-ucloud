---
layout: "ucloud"
page_title: "UCloud: ucloud_nat_gateway"
sidebar_current: "docs-ucloud-resource-nat-gateway"
description: |-
  Provides a Nat Gateway resource.
---

# ucloud_nat_gateway

Provides a Nat Gateway resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-nat-gateway-basic"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-nat-gateway-basic"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-nat-gateway-basic"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
  type = "recommend_web"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateway-basic"
  tag               = "tf-acc"
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the Nat Gateway. 
* `subnet_ids` - (Required) The list of subnet ID under the VPC.
* `eip_id` - (Required, ForceNew) The ID of eip associate to the Nat Gateway. 
* `security_group` - (Required) The ID of the associated security group.
* `enable_white_list` - (Required) The boolean value to Controls whether or not start the whitelist mode.

- - -

* `white_list` - (Optional) The white list of instance under the Nat Gateway.
* `name` - (Optional, ForceNew) The name of the Nat Gateway which contains 6-63 characters and only support Chinese, English, numbers, '-', '_' and '.'. If not specified, terraform will auto-generate a name beginning with `tf-nat-gateway-`.
* `remark` - (Optional, ForceNew) The remarks of the Nat Gateway. (Default: `""`).
* `tag` - (Optional, ForceNew) A tag assigned to Nat Gateway, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* ``
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of Nat Gateway, formatted in RFC3339 time string.

## Import

Nat Gateway can be imported using the `id`, e.g.

```
$ terraform import ucloud_nat_gateway.example natgw-abc123456
```