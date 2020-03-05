---
layout: "ucloud"
page_title: "UCloud: ucloud_vpn_connections"
sidebar_current: "docs-ucloud-datasource-vpn-connections"
description: |-
  Provides a list of VPN Connection resources in the current region.
---

# ucloud_vpn_connections

This data source providers a list of VPN Connection resources according to their ID, name and tag.

## Example Usage

```hcl
data "ucloud_vpn_connections" "example" {
}

output "first" {
  value = data.ucloud_vpn_connections.example.vpn_connections[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of VPN Connection IDs, all the VPN Connections belongs to the defined region will be retrieved if this argument is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting VPN Connections by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `tag` - (Optional) A tag assigned to VPN Connection.
* `vpc_id` - (Optional) The ID of VPC linked to the VPN Connection.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpn_connections` - It is a nested type. VPN Connections documented below.
* `total_count` - Total number of VPN Connections that satisfy the condition.

- - -

The attribute (`vpn_connections`) support the following:

* `id` - The ID of VPN Connection.
* `name` - The name of the VPN Connection.
* `remark` - The remarks of VPN Connection.
* `tag` - A tag assigned to the VPN Connection.
* `vpn_gateway_id` - The ID of VPN Gateway.
* `customer_gateway_id` - The ID of VPN Customer Gateway.
* `vpc_id` - The ID of VPC linked to the VPN Connection.
* `create_time` - The time of creation for VPN Connection, formatted in RFC3339 time string.
* `ike_config` - It is a nested type which documented below.
* `ipsec_config` - It is a nested type which documented below.

The attribute (`ike_config`) supports the following:

* `pre_shared_key` - The key used for authentication between the VPN gateway and the Customer gateway.
* `ike_version` - The version of the IKE protocol.
* `exchange_mode` - The negotiation exchange mode of IKE V1 of VPN gateway. 
* `encryption_algorithm` - The encryption algorithm of IKE negotiation.
* `authentication_algorithm` - The authentication algorithm of IKE negotiation.
* `local_id` - The identification of the VPN gateway.
* `remote_id` - The identification of the Customer gateway.
* `dh_group` - The Diffie-Hellman group used by IKE negotiation.
* `sa_life_time` - The Security Association lifecycle as the result of IKE negotiation.

The attribute (`ipsec_config`) supports the following:

* `local_subnet_ids` - The id list of Local subnet. 
* `remote_subnets` - The ip address list of remote subnet.
* `protocol` - The security protocol of IPSec negotiation.
* `encryption_algorithm` - The encryption algorithm of IPSec negotiation.
* `authentication_algorithm` - The authentication algorithm of IPSec negotiation.
* `pfs_dh_group` - Whether the PFS of IPSec negotiation is on or off, `disable` as off, The Diffie-Hellman group as open.
* `sa_life_time` - The Security Association lifecycle as the result of IPSec negotiation.
* `sa_life_time_bytes` - The Security Association lifecycle in bytes as the result of IPSec negotiation.