---
layout: "ucloud"
page_title: "UCloud: ucloud_lb"
sidebar_current: "docs-ucloud-resource-lb"
description: |-
  Provides a Load Balancer resource.
---

# ucloud_lb

Provides a Load Balancer resource.

## Example Usage

```hcl
resource "ucloud_lb" "web" {
    name = "tf-example-lb"
    tag  = "tf-example"
}
```

## Argument Reference

The following arguments are supported:

* `internal` - (Optional) Indicate whether the LB is intranet.
* `internet_charge_type` - (Optional) Charge type of LB. Possible values are: "Year" as pay by year, "Month" as pay by month, "Dynamic" as pay by hour (specific permission required). The default value is "Month".
* `name` - (Optional) The name of the load balancer, default is "LB".
* `remark` - (Optional) The remarks of the LB, the default value is "".
* `subnet_id` - (Optional) The ID of subnet that intrant LB belongs to, This argumnet is not required if default subnet.
* `tag` - (Optional) A mapping of tags to assign to the load balancer., the default value is "Default"(means no tag assigned).
* `vpc_id` - (Optional) ID of the VPC linked to the LBs, This argumnet is not required if default VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation for LB.
* `expire_time` - The expiration time for LB.
* `ip_set` - ip_set is a nested type. ip_set documented below.
* `private_ip` - The IP address of intranet IP, this attribute is "" if "internal" is "false".

The attribute (`ip_set`) support the following:

* `eip_id` - The ID of EIP.
* `internet_type` - Elastic IP routes. Possible values are: "International" as internaltional IP and "Bgp" as BGP IP.
* `ip` - Elastic IP address.
