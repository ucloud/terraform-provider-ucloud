---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_network_interface"
description: |-
  Provides a virtual network interface (UNI) resource.
---

# ucloud_network_interface

Provides a virtual network interface (UNI) resource. A UNI is an elastic network interface that can be created in a
subnet and optionally attached to a UHost instance. Detaching or deleting the resource leaves the instance untouched.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-example-subnet"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_network_interface" "foo" {
  name      = "tf-example-uni"
  tag       = "tf-example"
  vpc_id    = ucloud_vpc.foo.id
  subnet_id = ucloud_subnet.foo.id

  # Optionally attach the interface to an instance.
  # instance_id = "uhost-xxxx"
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, ForceNew) The ID of the VPC the network interface belongs to.
* `subnet_id` - (Required, ForceNew) The ID of the subnet the network interface belongs to.
* `name` - (Optional) The name of the network interface. If not specified, a name is auto generated.
* `tag` - (Optional) The business group the network interface belongs to. Defaults to `Default`.
* `remark` - (Optional) The remark of the network interface.
* `private_ip` - (Optional, ForceNew) The private IP address to assign to the network interface. If not specified,
  one is allocated automatically.
* `instance_id` - (Optional) The ID of the UHost instance to attach the network interface to. Setting this attaches
  the interface; removing it detaches the interface.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the network interface.
* `mac_address` - The MAC address of the network interface.
* `status` - The binding status of the network interface. `1` means attached to an instance, `0` means detached.
* `create_time` - The time when the network interface was created.

## Import

Network Interface can be imported using the interface id, e.g.

```
terraform import ucloud_network_interface.example uni-abc123456
```
