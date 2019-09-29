---
layout: "ucloud"
page_title: "UCloud: ucloud_vpc_peering_connection"
sidebar_current: "docs-ucloud-resource-vpc-peering-connection"
description: |-
  Provides an VPC Peering Connection for establishing a connection between multiple VPC.
---

# ucloud_vpc_peering_connection

Provides an VPC Peering Connection for establishing a connection between multiple VPC.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpc-01"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_vpc" "bar" {
  name        = "tf-example-vpc-02"
  tag         = "tf-example"
  cidr_blocks = ["10.10.0.0/16"]
}

resource "ucloud_vpc_peering_connection" "connection" {
  vpc_id      = ucloud_vpc.foo.id
  peer_vpc_id = ucloud_vpc.bar.id
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The short of ID of the requester VPC of the specific VPC Peering Connection to retrieve.
* `peer_vpc_id` - (Required) The short ID of accepter VPC of the specific VPC Peering Connection to retrieve.

- - -

* `peer_project_id` - (Optional) The ID of accepter project of the specific VPC Peering Connection to retrieve.