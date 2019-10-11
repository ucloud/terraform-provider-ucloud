---
layout: "ucloud"
page_title: "UCloud: ucloud_vpn_gateway"
sidebar_current: "docs-ucloud-resource-vpn-gateway"
description: |-
  Provides a VPN Gateway resource.
---

# ucloud_vpn_gateway

Provides a VPN Gateway resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpn-gateway-basic"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-vpn-gateway-basic"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
	vpc_id	 	= ucloud_vpc.foo.id
	grade		= "enhanced"
	eip_id		= ucloud_eip.foo.id
	name 		= "tf-acc-vpn-gateway-basic"
	tag         = "tf-acc"
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of VPC linked to the VPN Gateway. 
* `grade` - (Required) The type of the VPN Gateway. Possible values: `standard`, `enhanced`. `standard` recommended application scenario: Applicable to services with bidirectional peak bandwidth of 1M~50M; `enhanced` recommended application scenario: Suitable for services with bidirectional peak bandwidths of 50M~100M.
* `eip_id` - (Required) The ID of eip associate to the VPN Gateway. 
* `security_group` - (Required) The ID of the associated security group.

- - -

* `charge_type` - (Optional) The charge type of VPN Gateway, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional) The duration that you will buy the VPN Gateway (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `name` - (Optional) The name of the VPN Gateway which contains 1-63 characters and only support Chinese, English, numbers, '-', '_' and '.'. If not specified, terraform will auto-generate a name beginning with `tf-vpn-gateway-`.
* `remark` - (Optional) The remarks of the VPN Gateway. (Default: `""`).
* `tag` - (Optional) A tag assigned to VPN Gateway, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* ``
## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The creation time for VPN Gateway, formatted in RFC3339 time string.
* `expire_time` - The expiration time for VPN Gateway, formatted in RFC3339 time string.

## Import

VPN Gateway can be imported using the `id`, e.g.

```
$ terraform import ucloud_vpn_gateway.example vpngw-abc123456
```