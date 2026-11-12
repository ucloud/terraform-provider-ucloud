---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_sec_group"
description: |-
  Provides a Security Group (SecGroup) resource.
---

# ucloud_sec_group

Provides a Security Group resource, which belongs to a VPC.

~> **Note** This resource manages the VPC `SecGroup` API, not the `Firewall` API behind
[`ucloud_security_group`](./security_group.html.markdown). The two are different objects with
different IDs and different rule models, and they are not interchangeable.

~> **Note** Rules are not part of this resource. Each one is managed by
[`ucloud_sec_group_rule`](./sec_group_rule.html.markdown), so that a rule can be changed in place,
imported and destroyed on its own.

## Example Usage

```hcl
resource "ucloud_vpc" "example" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "example" {
  name   = "tf-example-sec-group"
  vpc_id = ucloud_vpc.example.id
  remark = "managed by terraform"
}

resource "ucloud_sec_group_rule" "example" {
  sec_group_id  = ucloud_sec_group.example.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = "22"
  ip_range      = "10.0.0.0/8"
  rule_action   = "Accept"
  priority      = 50
  remark        = "ssh"
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, ForceNew) The ID of the VPC the security group belongs to.

- - -

* `name` - (Optional) The name of the security group. If not specified, terraform will auto-generate a name beginning with `tf-sec-group`.
* `remark` - (Optional) The remarks of the security group. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the security group.
* `create_time` - The time of creation for the security group, formatted in RFC3339 time string.
* `type` - The type of the security group.
* `tag` - A tag assigned to the security group.

~> **Note** `tag` is read-only. Neither the create nor the update API of a security group accepts a
tag, so it is decided by the platform and cannot be set from a configuration.

## Import

Security Group can be imported using the `id`, e.g.

```
$ terraform import ucloud_sec_group.example secgroup-abc123456
```
