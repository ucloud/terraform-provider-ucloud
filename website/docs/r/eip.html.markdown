---
layout: "ucloud"
page_title: "UCloud: ucloud_eip"
sidebar_current: "docs-ucloud-resource-eip"
description: |-
  Provides an Elastic IP resource.
---

# ucloud_eip

Provides an Elastic IP resource.

## Example Usage

```hcl
resource "ucloud_eip" "example" {
    bandwidth            = 2
    charge_mode          = "bandwidth"
    name                 = "tf-example-eip"
    tag                  = "tf-example"
    internet_type        = "bgp"
}
```

## Argument Reference

The following arguments are supported:

* `internet_type` - (Required) Type of Elastic IP routes. Possible values are: `international` as internaltional BGP IP and `bgp` as china BGP IP.
* `bandwidth` - (Optional) Maximum bandwidth to the elastic public network, measured in Mbps (Mega bit per second). the ranges for bandwidth are: 1-200 for pay by traffic, 1-800 for pay by bandwith. (Default: `1`).
* `duration` - (Optional) The duration that you will buy the resource. (Default: `1`). It is not required when `dynamic` (pay by hour), the value is `0` when `month`(pay by month) and the instance will be vaild till the last day of that month.
* `charge_mode` -(Optional) Elastic IP charge mode. Possible values are: `traffic` as pay by traffic, `bandwidth` as pay by bandwidth. (Default: `bandwidth`).
* `charge_type` - (Optional) Elastic IP charge type. Possible values are: `year` as pay by year, `month` as pay by month, `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `name` - (Optional) The name of the EIP, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will autogenerate a name beginning with `tf-eip`.
* `remark` - (Optional) The remarks of the EIP. (Default: `""`).
* `tag` - (Optional) A mapping of tags to assign to VPC, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation for EIP, formatted in RFC3339 time string.
* `expire_time` - The expiration time for EIP, formatted in RFC3339 time string.
* `ip_set` - It is a nested type which documented below.
* `resource` - It is a nested type which documented below.
* `status` - EIP status. Possible values are: `used` as in use, `free` as available and `freeze` as associating.

The attribute (`ip_set`) support the following:

* `internet_type` - Type of Elastic IP routes.

The attribute (`resource`) support the following:

* `id` - The ID of the resource with EIP attached.
* `type` - The type of resource with EIP attached. Possible values are `instance` as instance, `vrouter` as visual router, `lb` as load balancer.
