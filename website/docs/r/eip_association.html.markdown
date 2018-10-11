---
layout: "ucloud"
page_title: "UCloud: ucloud_eip_association"
sidebar_current: "docs-ucloud-resource-eip-association"
description: |-
  Provides an EIP Association resource for associating Elastic IP to UHost Instance, Load Balancer, etc..
---

# ucloud_eip_association

Provides an EIP Association resource for associating Elastic IP to UHost Instance, Load Balancer, etc.

## Example Usage

```hcl
resource "ucloud_security_group" "default" {
    name = "tf-example-eip"
    tag  = "tf-example"

    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

resource "ucloud_eip" "default" {
    bandwidth            = 2
    internet_charge_mode = "Bandwidth"
    name                 = "tf-example-eip"
    tag                  = "tf-example"
}

resource "ucloud_instance" "web" {
    instance_type     = "n-standard-1"
    availability_zone = "cn-sh2-02"

    root_password      = "wA1234567"
    image_id           = "uimage-of3pac"
    security_group     = "${ucloud_security_group.default.id}"

    name              = "tf-example-eip"
    tag               = "tf-example"
}

resource "ucloud_eip_association" "default" {
    resource_type = "instance"
    resource_id   = "${ucloud_instance.web.id}"
    eip_id        = "${ucloud_eip.default.id}"
}
```

## Argument Reference

The following arguments are supported:

* `eip_id` - (Required) The ID of EIP.
* `resource_id` - (Required) The ID of resource with EIP attached.
* `resource_type` - (Required) The type of resource with EIP attached, possible values are "instance" as instance, "vrouter" as virtual router, "lb" as load balancer, "upm" as physical server, "hadoophost" as hadoop cluster, "fortresshost" as fortress host server, "udockhost" as docker host, "udhost" as dedicated host, "natgw" as NAT GateWay host, "udb" as data base host, "vpngw" as ipsec vpn host, "ucdr" as cloud diaster recovery host, "dbaudit" as data base auditing host.