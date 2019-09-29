---
layout: "ucloud"
page_title: "UCloud: ucloud_security_group"
sidebar_current: "docs-ucloud-resource-security-group"
description: |-
  Provides a Security Group resource.
---

# ucloud_security_group

Provides a Security Group resource.

## Example Usage

```hcl
resource "ucloud_security_group" "example" {
  name = "tf-example-security-group"
  tag  = "tf-example"

  # http access from LAN
  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }

  # https access from LAN
  rules {
    port_range = "443"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }
}
```

## Argument Reference

The following arguments are supported:

* `rules` - (Required) A list of security group rules. Can be specified multiple times for each rules. Each rules supports fields documented below.

- - -

* `name` - (Optional) The name of the security group which contains 1-63 characters and only support Chinese, English, numbers, '-', '_' and '.'. If not specified, terraform will auto-generate a name beginning with `tf-security-group`.
* `remark` - (Optional) The remarks of the security group. (Default: `""`).
* `tag` - (Optional) A tag assigned to security group, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

### Block rules

The rules mapping supports the following:

* `port_range` - (Optional) The range of port numbers, range: 1-65535. (eg: `port` or `port1-port2`).
* `cidr_block` - (Optional) The cidr block of source.
* `policy` - (Optional) Authorization policy. Possible values are: `accept`, `drop`.
* `priority` - (Optional) Rule priority. Possible values are: `high`, `medium`, `low`.
* `protocol` - (Optional) The protocol. Possible values are: `tcp`, `udp`, `icmp`, `gre`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of security group, formatted in RFC3339 time string.

## Import

Security Group can be imported using the `id`, e.g.

```
$ terraform import ucloud_security_group.example firewall-abc123456
```