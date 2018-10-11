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

    # HTTP access from LAN
    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }

    # HTTPS access from LAN
    rules {
        port_range = "443"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}
```

## Argument Reference

The following arguments are supported:

* `rules` - (Required) A list of security group rules. Each element contains the following attributes: protocol, port, ip, policy ("ACCEPT" and "DROP") and priority ("HIGH", "MEDIUM" and "LOW". ( TCP|22|192.168.1.1/22|DROP|LOW etc.)
* `name` - (Optional) The name of the security group, default is "SecurityGroup", should have 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'.
* `remark` - (Optional) The remarks of the security group, the default value is "".
* `tag` - (Optional) A mapping of tags to assign to the security group, the default value is"Default"(means no tag assigned).

The attribute (`rules`) support the following:

* `cidr_block` - The cidr block of source.
* `policy` - Authorization policy. Can be either ACCEPT or DROP.
* `port_range` - The range of port numbers, any port should between 1-65535, should like a port or port1-port2.
* `priority` - Rule priority. Can be HIGH, MEDIUM, LOW.
* `protocol` - The protocol. Can be TCP, UDP, ICMP, GRE.
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of security group.
