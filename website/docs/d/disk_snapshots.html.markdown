---
subcategory: "UHost"
layout: "ucloud"
page_title: "UCloud: ucloud_disk_snapshots"
description: |-
  Provides a list of Disk Snapshot resources in the current region.
---

# ucloud_disk_snapshots

This data source provides a list of Disk Snapshot resources according to their availability zone, source disk ID, snapshot ID and name.

## Example Usage

```hcl
data "ucloud_disk_snapshots" "example" {
  availability_zone = "cn-bj2-02"
}

output "first" {
  value = data.ucloud_disk_snapshots.example.snapshots[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where the snapshots are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist). It is required when `disk_id` is set.
* `disk_id` - (Optional) The ID of the cloud disk which the snapshots are created from. The `availability_zone` must be set at the same time.
* `ids` - (Optional) A list of snapshot IDs, all the snapshots belong to this region will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter the resulting snapshots by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `snapshots` - It is a nested type which documented below.
* `total_count` - Total number of snapshots that satisfy the condition.

The attribute (`snapshots`) supports the following:

* `id` - The ID of the snapshot.
* `availability_zone` - Availability zone where the snapshot is located.
* `disk_id` - The ID of the cloud disk which the snapshot is created from.
* `name` - The name of the snapshot.
* `comment` - The description of the snapshot.
* `disk_type` - The type of the source disk. Possible values are: `data_disk` as cloud disk, `ssd_data_disk` as ssd cloud disk, `rssd_data_disk` as RDMA-SSD cloud disk, `system_disk` as cloud system disk, `ssd_system_disk` as ssd cloud system disk.
* `size` - The size of the snapshot in GB.
* `source_disk_name` - The name of the cloud disk which the snapshot is created from.
* `is_disk_available` - Whether the source disk of the snapshot is available.
* `create_time` - The time of creation of the snapshot, formatted in RFC3339 time string.
* `status` - The status of the snapshot. Possible values are: `Normal`, `Failed`, `Creating`.
