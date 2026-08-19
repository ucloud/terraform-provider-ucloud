---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_sec_group"
description: |-
  Provides a Security Group (SecGroup) resource.
---

# ucloud_sec_group

Provides a Security Group (SecGroup) resource in a VPC.

~> **Note** `ucloud_sec_group` and `ucloud_security_group` are two different products. `ucloud_sec_group` is the VPC security group, while `ucloud_security_group` is the firewall. They are not interchangeable, and a resource created by one of them can not be managed by the other.

~> **Note** The rules of a security group are managed by the separate [`ucloud_sec_group_rule`](/docs/providers/ucloud/r/sec_group_rule.html) resource. The `rules` attribute exported here is read only, it is a snapshot of the rules that currently belong to the group, including the ones created outside of terraform.

## Example Usage

```hcl
resource "ucloud_vpc" "example" {
  name        = "tf-example-vpc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "example" {
  name   = "tf-example-sec-group"
  vpc_id = ucloud_vpc.example.id
  tag    = "tf-example"
  remark = "managed by terraform"
}

resource "ucloud_sec_group_rule" "ssh" {
  sec_group_id  = ucloud_sec_group.example.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = "22"
  ip_range      = "192.168.0.0/16"
  rule_action   = "Accept"
  priority      = 50
}
```

Bind the security group to an instance:

```hcl
resource "ucloud_instance" "example" {
  # ... other configuration ...

  security_mode = "SecGroup"

  sec_group_id {
    id       = ucloud_sec_group.example.id
    priority = 1
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the security group, should have 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'.
* `vpc_id` - (Required, ForceNew) The ID of the VPC that the security group belongs to. The remote api does not support moving a security group between VPCs, so any change of it will destroy the existing security group and create a new one.

- - -

* `tag` - (Optional) A tag assigned to the security group, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `remark` - (Optional) The remarks of the security group. The remote api ignores an empty value instead of clearing the field, so an existing remark can be replaced but not removed.

~> **Note** The remote api accepts neither `tag` nor `remark` while the security group is being created, both are applied right after the creation. A failure in that second step leaves the security group tainted, and the next apply recreates it.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the security group.
* `type` - The type of the security group. Possible values are: `user defined` as a user defined security group, `recommend web` as a security group created from the web template, `recommend non web` as a security group created from the non web template.
* `create_time` - The time of creation, formatted in RFC3339 time string.
* `rules` - A read only list of the rules that currently belong to the security group. Each element contains the following attributes:
    * `rule_id` - The ID of the rule.
    * `direction` - The direction of the rule. Possible values are: `Ingress`, `Egress`.
    * `protocol_type` - The protocol type of the rule. Possible values are: `TCP`, `UDP`, `ICMP`, `ICMPv6`, `ALL`.
    * `dst_port` - The destination port of the rule.
    * `ip_range` - The IP address range of the rule.
    * `rule_action` - The action of the rule. Possible values are: `Accept`, `Drop`.
    * `priority` - The priority of the rule.
    * `remark` - The remarks of the rule.

## Import

Security group can be imported using the `id`, e.g.

```
$ terraform import ucloud_sec_group.example secgroup-abcdefg
```
