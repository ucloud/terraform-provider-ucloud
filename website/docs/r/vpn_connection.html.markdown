---
layout: "ucloud"
page_title: "UCloud: ucloud_vpn_connection"
sidebar_current: "docs-ucloud-resource-vpn-connection"
description: |-
  Provides a IPSec VPN Gateway Connection resource.
---

# ucloud_vpn_connection

Provides a IPSec VPN Gateway Connection resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-vpn-connection-basic"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-vpn-connection-basic"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-vpn-connection-basic"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

resource "ucloud_vpn_gateway" "foo" {
  vpc_id = ucloud_vpc.foo.id
  grade  = "standard"
  eip_id = ucloud_eip.foo.id
  name   = "tf-acc-vpn-connection-basic"
  tag    = "tf-acc"
}

resource "ucloud_vpn_customer_gateway" "foo" {
  ip_address = "10.0.0.1"
  name       = "tf-acc-vpn-connection-basic"
  tag        = "tf-acc"
}

resource "ucloud_vpn_connection" "foo" {
  vpn_gateway_id      = ucloud_vpn_gateway.foo.id
  customer_gateway_id = ucloud_vpn_customer_gateway.foo.id
  vpc_id              = ucloud_vpc.foo.id
  name                = "tf-acc-vpn-connection-basic"
  tag                 = "tf-acc"
  remark              = "test"
  ike_config {
    pre_shared_key = "test_2019"
  }

  ipsec_config {
    local_subnet_ids = [ucloud_subnet.foo.id]
    remote_subnets   = ["10.0.0.0/24"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of VPC linked to the VPN Gateway Connection. 
* `vpn_gateway_id` - (Required) The ID of  the VPN Customer Gateway. 
* `customer_gateway_id` - (Required) The grade of the VPN Gateway
* `ike_config` - (Required) The configurations of IKE negotiation. Each ike_config supports fields documented below.
* `ipsec_config` - (Required) The configurations of IPSec negotiation. Each ipsec_config supports fields documented below.

- - -

* `name` - (Optional) The name of the VPN Gateway Connection which contains 1-63 characters and only support Chinese, English, numbers and special characters: `-_.`. If not specified, terraform will auto-generate a name beginning with `tf-vpn-connection-`.
* `remark` - (Optional) The remarks of the VPN Gateway Connection. (Default: `""`).
* `tag` - (Optional) A tag assigned to VPN Gateway Connection, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

### Block ike_config

The ike_config mapping supports the following:

* `pre_shared_key` - (Required) The key used for authentication between the VPN gateway and the Customer gateway which contains 1-128 characters and only support English, numbers and special characters: `!@#$%^&*()_+-=[]:,./'~`.
* `ike_version` - (Optional) The version of the IKE protocol which only be supported IKE V1 protocol at present. Possible values: ikev1. (Default: ikev1)
* `exchange_mode` - (Optional) The negotiation exchange mode of IKE V1 of VPN gateway. Possible values: `main` (main mode), `aggressive` (aggressive mode). (Default: `main`)
* `encryption_algorithm` - (Optional) The encryption algorithm of IKE negotiation. Possible values: `aes128`, `aes192`, `aes256`, `aes512`, `3des`. (Default: `aes128`).
* `authentication_algorithm` - (Optional) The authentication algorithm of IKE negotiation. Possible values: `sha1`, `md5`, `sha2-256`. (Default: `sha1`)
* `local_id` - (Optional) The identification of the VPN gateway.
* `remote_id` - (Optional) The identification of the Customer gateway.
* `dh_group` - (Optional) The Diffie-Hellman group used by IKE negotiation. Possible values: `1`, `2`, `5`, `14`, `15`, `16`. (Default:`15`)
* `sa_life_time` - (Optional) The Security Association lifecycle as the result of IKE negotiation. Unit: second. Range: 600-604800. (Default: `86400`)


### Block ipsec_config

The ipsec_config mapping supports the following:

* `local_subnet_ids` - (Required) The id list of Local subnet. 
* `remote_subnets` - (Required) The ip address list of remote subnet.
* `protocol` - (Optional) The security protocol of IPSec negotiation. Possible values: `esp`, `ah`. (Default:`esp`)
* `encryption_algorithm` - (Optional) The encryption algorithm of IPSec negotiation. Possible values: `aes128`, `aes192`, `aes256`, `aes512`, `3des`. (Default: `aes128`).
* `authentication_algorithm` - (Optional) The authentication algorithm of IPSec negotiation. Possible values: `sha1`, `md5`. (Default: `sha1`)
* `pfs_dh_group` - (Optional) Whether the PFS of IPSec negotiation is on or off, `disable` as off, The Diffie-Hellman group as open.  Possible values: `disable`, `1`, `2`, `5`, `14`, `15`, `16`. (Default:`disable`)
* `sa_life_time` - (Optional) The Security Association lifecycle as the result of IPSec negotiation. Unit: second. Range: 1200-604800. (Default: `3600`)
* `sa_life_time_bytes` - (Optional) The Security Association lifecycle in bytes as the result of IPSec negotiation. Unit: second. Range: 1200-604800. (Default: `3600`)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The creation time for VPN Gateway Connection, formatted in RFC3339 time string.
* `expire_time` - The expiration time for VPN Gateway Connection, formatted in RFC3339 time string.