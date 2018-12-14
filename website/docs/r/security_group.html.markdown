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
    name = "tf-example-instance"
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

* `rules` - (Required) A list of security group rules. Each element contains the following attributes: `protocol`, `port_range`, `cidr_block`, `policy` (possbile values are:`"accept"` and `"drop"`) and priority (possible values are: `"high"`, `"medium"` and `"low"`. (eg: tcp|22|192.168.1.1/22|drop|low).
* `name` - (Optional) The name of the security group which contains 1-63 characters and only support Chinese, English, numbers, '-', '_' and '.'. If not specified, terraform will autogenerate a name beginning with `"tf-security-group"`.
* `remark` - (Optional) The remarks of the security group. (Default: `""`).
* `tag` - (Optional) A mapping of tags to assign to the security group,  which contains 1-63 characters and only support Chinese, English, numbers, '-', '_' and '.'. (Default: `"Default"`).

The attribute (`rules`) support the following:

* `cidr_block` - The cidr block of source.
* `policy` - Authorization policy. Can be either `"accept"` or `"drop"`.
* `port_range` - The range of port numbers, range: 1-65535. (eg: `"port"` or `"port1-port2"`).
* `priority` - Rule priority. Can be `"high"`, `"medium"`, `"low"`.
* `protocol` - The protocol. Can be `"tcp"`, `"udp"`, `"icmp"`, `"gre"`.
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of security group, formatted in RFC3339 time string.
