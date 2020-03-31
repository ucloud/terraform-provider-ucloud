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

* `internal` - (Optional, ForceNew) Indicate whether the load balancer is intranet mode.(Default: `"false"`)
* `name` - (Optional) The name of the load balancer. If not specified, terraform will auto-generate a name beginning with `tf-lb`.
* `charge_type` - (**Deprecated**, ForceNew), argument `charge_type` is deprecated for optimizing parameters.
* `vpc_id` - (Optional, ForceNew) The ID of the VPC linked to the Load balancer, This argument is not required if default VPC.
* `subnet_id` - (Optional, ForceNew) The ID of subnet that intranet load balancer belongs to. This argument is not required if default subnet.
* `tag` - (Optional) A tag assigned to load balancer, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `remark` - (Optional) The remarks of the load balancer. (Default: `""`).
* `security_group` - (Optional) The ID of the associated security group. The security_group only takes effect for ULB instances of request_proxy mode and extranet mode at present.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation for load balancer, formatted in RFC3339 time string.
* `expire_time` - **Deprecated** attribute `expire_time` is deprecated for optimizing outputs.
* `ip_set` - It is a nested type which documented below.
* `private_ip` - The IP address of intranet IP. It is `""` if `internal` is `false`.

The attribute (`ip_set`) support the following:

* `internet_type` - Type of Elastic IP routes.
* `ip` - Elastic IP address.

## Import

LB can be imported using the `id`, e.g.

```
$ terraform import ucloud_lb.example ulb-abc123456
```