---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_sec_group_rule"
description: |-
  Provides a Security Group (SecGroup) rule resource.
---

# ucloud_sec_group_rule

Provides a rule of a Security Group (SecGroup). Each resource represents a single ingress or egress rule of the security group referenced by `sec_group_id`.

~> **Note** This resource only manages the rules it declares. Rules added to the same security group outside of terraform are left untouched, they show up in the read only `rules` attribute of [`ucloud_sec_group`](/docs/providers/ucloud/r/sec_group.html).

~> **Note** This resource belongs to the VPC security group, not to the firewall. The rules of a firewall are declared inline in the `rules` block of `ucloud_security_group`, and the two are not interchangeable.

## Example Usage

```hcl
resource "ucloud_vpc" "example" {
  name        = "tf-example-vpc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "example" {
  name   = "tf-example-sec-group"
  vpc_id = ucloud_vpc.example.id
}

resource "ucloud_sec_group_rule" "ssh" {
  sec_group_id  = ucloud_sec_group.example.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = "22"
  ip_range      = "192.168.0.0/16"
  rule_action   = "Accept"
  priority      = 50
  remark        = "ssh from intranet"
}
```

Generate several rules from a map:

```hcl
variable "ingress_rules" {
  type = map(object({
    dst_port = string
    ip_range = string
    priority = number
  }))

  default = {
    http  = { dst_port = "80", ip_range = "0.0.0.0/0", priority = 50 }
    https = { dst_port = "443", ip_range = "0.0.0.0/0", priority = 50 }
    mysql = { dst_port = "3306", ip_range = "192.168.0.0/16", priority = 60 }
  }
}

resource "ucloud_sec_group_rule" "ingress" {
  for_each = var.ingress_rules

  sec_group_id  = ucloud_sec_group.example.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = each.value.dst_port
  ip_range      = each.value.ip_range
  rule_action   = "Accept"
  priority      = each.value.priority
  remark        = each.key
}
```

## Argument Reference

The following arguments are supported:

* `sec_group_id` - (Required, ForceNew) The ID of the security group that the rule belongs to. A rule can not be moved between security groups, so any change of it will destroy the existing rule and create a new one.
* `direction` - (Required) The direction of the rule. Possible values are: `Ingress` as an inbound rule, `Egress` as an outbound rule.
* `protocol_type` - (Required) The protocol type of the rule. Possible values are: `TCP`, `UDP`, `ICMP`, `ICMPv6`, `ALL`.
* `dst_port` - (Required) The destination port of the rule, separated by commas. Such as: `80`, `80,443`, `443,2000-10000`.
* `ip_range` - (Required) The IP address range of the rule, separated by commas. Such as: `0.0.0.0/0`, `192.168.0.0/16`.
* `rule_action` - (Required) The action taken when the rule matches. Possible values are: `Accept`, `Drop`.
* `priority` - (Required) The priority of the rule, ranges from 1 to 200.

- - -

* `remark` - (Optional) The remarks of the rule. Defaults to an empty string.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the rule, formatted as `<sec_group_id>:<rule_id>`.

## Import

Security group rule can be imported using the `id`, e.g.

```
$ terraform import ucloud_sec_group_rule.example secgroup-abcdefg:sgrule-hijklmn
```
