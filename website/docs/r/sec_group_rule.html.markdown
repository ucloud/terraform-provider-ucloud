---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_sec_group_rule"
description: |-
  Provides a Security Group Rule resource.
---

# ucloud_sec_group_rule

Provides a rule of a [security group](./sec_group.html.markdown). One resource manages exactly one
rule, which is what lets a rule be changed in place, imported and destroyed without touching the
rest of the group.

## Example Usage

```hcl
resource "ucloud_sec_group_rule" "example" {
  sec_group_id  = ucloud_sec_group.example.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = "80,443"
  ip_range      = "0.0.0.0/0"
  rule_action   = "Accept"
  priority      = 60
  remark        = "web"
}
```

## Argument Reference

The following arguments are supported:

* `sec_group_id` - (Required, ForceNew) The ID of the security group the rule belongs to, e.g. `secgroup-abc123`. The group is the rule's parent rather than one of its attributes, so changing it rebuilds the rule.
* `direction` - (Required) The direction of the rule. Possible values are: `Ingress`, `Egress`.
* `protocol_type` - (Required) The protocol type. Possible values are: `TCP`, `UDP`, `ICMP`, `ICMPv6`, `ALL`.
* `ip_range` - (Required) The source the rule applies to, in one of three mutually exclusive forms: a comma separated list of CIDR blocks (e.g. `10.0.0.0/8` or `10.0.0.0/8,192.168.0.0/16`), a single security group ID (e.g. `secgroup-abc123`), or a comma separated list of prefix list IDs (e.g. `pl-abc123,pl-def456`). The forms cannot be mixed. Whatever the form carries, it is one value of a single rule rather than a way to add several: exactly one rule is created.
* `rule_action` - (Required) The action of the rule. Possible values are: `Accept`, `Drop`.
* `priority` - (Required) The priority of the rule, between `1` and `200`. A smaller value means a higher priority.

- - -

* `dst_port` - (Optional) The destination port, a comma separated list where each item is a single port or a `from-to` range, e.g. `80,443` or `443,2000-10000`. Required when `protocol_type` is `TCP` or `UDP`, and must be omitted when it is `ICMP`, `ICMPv6` or `ALL`, where a port carries no meaning.
* `remark` - (Optional) The remarks of the rule. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the rule, assigned by the API.

~> **Note** Every argument except `sec_group_id` is edited in place: the API addresses a rule by the
ID it assigned, so changing a port, a range or the priority keeps the same rule and the same `id`.

## Import

Security Group Rule can be imported using the ID of its group and the ID of the rule, separated by a
slash, e.g.

```
$ terraform import ucloud_sec_group_rule.example secgroup-abc123456/rule-def789012
```

~> **Note** The group has to be part of the import ID because the API cannot look a rule up from its
own ID alone.
