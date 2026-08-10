---
subcategory: "UHost"
layout: "ucloud"
page_title: "UCloud: ucloud_disk_snapshot"
description: |-
  Provides a Cloud Disk Snapshot resource.
---

# ucloud_disk_snapshot

Provides a Cloud Disk Snapshot resource.

~> **Note** The snapshot service must have been enabled on the source cloud disk, otherwise the creation fails with `snapshot service not avaiable`. The snapshot service can only be enabled while the disk is being created, by setting `snapshot_service` of `ucloud_disk` to `true`. There is no way to enable it on an existing disk through terraform.

~> **Note** All the arguments of the snapshot are immutable, any change of them will destroy the existing snapshot and create a new one.

## Example Usage

```hcl
# Query availability zone
data "ucloud_zones" "default" {}

# Create cloud disk with the snapshot service enabled
resource "ucloud_disk" "example" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-example-disk"
  disk_size         = 20
  snapshot_service  = true
}

# Create a snapshot of the cloud disk
resource "ucloud_disk_snapshot" "example" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  disk_id           = ucloud_disk.example.id
  name              = "tf-example-disk-snapshot"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where the snapshot is located, it must be the same as the availability zone of the source disk. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist).
* `disk_id` - (Required, ForceNew) The ID of the cloud disk which the snapshot is created from.

- - -

* `name` - (Optional, ForceNew) The name of the snapshot, should have 6-63 characters and only support Chinese, English, numbers, '-', '_'. If not specified, terraform will auto-generate a name beginning with `tf-disk-snapshot`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the snapshot.
* `comment` - The description of the snapshot. It is read only, the remote api does not accept a description while the snapshot is being created.
* `disk_type` - The type of the source disk. Possible values are: `data_disk` as cloud disk, `ssd_data_disk` as ssd cloud disk, `rssd_data_disk` as RDMA-SSD cloud disk, `system_disk` as cloud system disk, `ssd_system_disk` as ssd cloud system disk.
* `size` - The size of the snapshot in GB.
* `source_disk_name` - The name of the cloud disk which the snapshot is created from.
* `is_disk_available` - Whether the source disk of the snapshot is available.
* `create_time` - The time of creation of the snapshot, formatted in RFC3339 time string.
* `status` - The status of the snapshot. Possible values are: `Normal`, `Failed`, `Creating`.

## Import

Disk snapshot can be imported using the `id`, e.g.

```
$ terraform import ucloud_disk_snapshot.example snap-abcdefg
```
