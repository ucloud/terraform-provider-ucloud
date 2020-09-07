---
subcategory: "IPSec VPN"
layout: "ucloud"
page_title: "UCloud: ucloud_vpn_gateways"
description: |-
  Provides a list of VPN Gateway resources in the current region.
---

# ucloud_vpn_gateways

This data source providers a list of VPN Gateway resources according to their ID, name, vpc and tag.

## Example Usage

```hcl
data "ucloud_vpn_gateways" "example" {
}

output "first" {
  value = data.ucloud_vpn_gateways.example.vpn_gateways[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of VPN Gateway IDs, all the VPN Gateways belongs to the defined region will be retrieved if this argument is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting VPN Gateways by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `tag` - (Optional) A tag assigned to VPN Gateway.
* `vpc_id` - (Optional) The ID of VPC linked to the VPN Gateway.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpn_gateways` - It is a nested type. VPN Gateways documented below.
* `total_count` - Total number of VPN Gateways that satisfy the condition.

- - -

The attribute (`vpn_gateways`) support the following:

* `id` - The ID of VPN Gateway.
* `name` - The name of the VPN Gateway.
* `remark` - The remarks of VPN Gateway.
* `tag` - A tag assigned to the VPN Gateway.
* `grade` - The type of the VPN Gateway.
* `vpc_id` - The ID of VPC linked to the VPN Gateway.
* `charge_type` - The charge type of VPN Gateway.
* `auto_renew` - Whether to renew an VPN Gateway automatically or not.
* `create_time` - The time of creation for VPN Gateway, formatted in RFC3339 time string.
* `expire_time` - The expiration time for VPN Gateway, formatted in RFC3339 time string.
* `ip_set` - It is a nested type which documented below.

The attribute (`ip_set`) supports the following:

* `internet_type` - Type of Elastic IP routes.
* `ip` - Elastic IP address.