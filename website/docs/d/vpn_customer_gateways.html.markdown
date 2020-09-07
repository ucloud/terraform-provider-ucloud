---
subcategory: "IPSec VPN"
layout: "ucloud"
page_title: "UCloud: ucloud_vpn_customer_gateways"
description: |-
  Provides a list of VPN Gateway resources in the current region.
---

# ucloud_vpn_customer_gateways

This data source providers a list of VPN Customer Gateway resources according to their ID, name and tag.

## Example Usage

```hcl
data "ucloud_vpn_customer_gateways" "example" {
}

output "first" {
  value = data.ucloud_vpn_customer_gateways.example.vpn_customer_gateways[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of VPN Customer Gateway IDs, all the VPN Customer Gateways belongs to the defined region will be retrieved if this argument is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting VPN Customer Gateways by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `tag` - (Optional) A tag assigned to VPN Customer Gateway.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpn_customer_gateways` - It is a nested type. VPN Customer Gateways documented below.
* `total_count` - Total number of VPN Customer Gateways that satisfy the condition.

- - -

The attribute (`vpn_customer_gateways`) support the following:

* `id` - The ID of VPN Customer Gateway.
* `name` - The name of the VPN Customer Gateway.
* `remark` - The remarks of VPN Customer Gateway.
* `tag` - A tag assigned to the VPN Customer Gateway.
* `ip_address` - The ip address of the VPN Customer Gateway.
* `create_time` - The time of creation for VPN Customer Gateway, formatted in RFC3339 time string.