---
subcategory: "UFS"
layout: "ucloud"
page_title: "UCloud: ucloud_ufs_volume_mount_point"
description: |-
  Provides a UFS Volume mount point resource.
---

# ucloud_ufs_volume

Provides a UFS Volume mount point resource.

## Example Usage

```hcl
data "ucloud_vpcs" "default" {
}

data "ucloud_subnets" "default" {
  vpc_id = data.ucloud_vpcs.default.vpcs[0].id
}

resource "ucloud_ufs_volume" "foo" {
  name          = "tf-acc-ufs-basic"
  remark        = "test"
  tag           = "tf-acc"
  size          = 600
  storage_type  = "Basic"
  protocol_type = "NFSv4"
}

resource "ucloud_ufs_volume_mount_point" "foo" {
  name      = "tf-acc-ufs-mount-point-basic"
  volume_id = ucloud_ufs_volume.foo.id
  vpc_id    = data.ucloud_vpcs.default.vpcs[0].id
  subnet_id = data.ucloud_subnets.default.subnets[0].id
}
```

## Argument Reference

The following arguments are supported:

* `volume_id` - (Required, ForceNew) The id of the UFS Volume.
* `name` - (Required, ForceNew) The name of the UFS Volume mount point, expected value to be 6 - 63 characters and only support english, numbers, '-', '_', and can not prefix with '-'.
* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the UFS Volume mount point.
* `subnet_id` - (Required, ForceNew) The ID of subnet.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource UFS Volume mount point.
* `mount_point_ip` - The ip of the UFS Volume mount point.
* `create_time` - The time of creation of the UFS Volume mount point, formatted in RFC3339 time string.