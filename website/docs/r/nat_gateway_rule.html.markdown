---
layout: "ucloud"
page_title: "UCloud: ucloud_nat_gateway_rule"
sidebar_current: "docs-ucloud-resource-nat-gateway-rule"
description: |-
  Provides a Nat Gateway Rule resource.
---

# ucloud_nat_gateway_rule

Provides a Nat Gateway resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-nat-gateway-rule-basic"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
  name       = "tf-acc-nat-gateway-rule-basic"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-nat-gateway-rule-basic"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
  type = "recommend_web"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateway-rule-basic"
  tag               = "tf-acc"
  enable_white_list = false
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}

resource "ucloud_nat_gateway_rule" "foo" {
  nat_gateway_id = ucloud_nat_gateway.foo.id
  protocol       = "tcp"
  src_eip_id     = ucloud_eip.foo.id
  src_port_range = "88"
  dst_ip         = ucloud_instance.foo.private_ip
  dst_port_range = "80"
  name           = "tf-acc-nat-gateway-rule-basic"
}

resource "ucloud_nat_gateway_rule" "bar" {
  nat_gateway_id = ucloud_nat_gateway.foo.id
  protocol       = "tcp"
  src_eip_id     = ucloud_eip.foo.id
  src_port_range = "90-100"
  dst_ip         = ucloud_instance.foo.private_ip
  dst_port_range = "90-100"
  name           = "tf-acc-nat-gateway-rule-basic"
}
```

## Argument Reference

The following arguments are supported:

* `nat_gateway_id` - (Required, ForceNew) The ID of the Nat Gateway. 
* `protocol` - (Required) The protocol of the Nat Gateway Rule. Possible values: `tcp`, `udp`.
* `src_eip_id` - (Required) The ID of eip associate to the Nat Gateway.
* `src_port_range` - (Required) The range of port numbers of the eip, range: 1-65535. (eg: `port` or `port1-port2`).
* `dst_ip` - (Required) The private ip of instance bound to the jNAT gateway.
* `dst_port_range` - (Required) The range of port numbers of the private ip, range: 1-65535. (eg: `port` or `port1-port2`).

- - -

* `name` - (Optional) The name of the Nat Gateway Rule which contains 6-63 characters and only support Chinese, English, numbers, '-', '_' and '.'. If not specified, terraform will auto-generate a name beginning with `tf-nat-gateway-rule-`.