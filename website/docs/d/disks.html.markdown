---
layout: "ucloud"
page_title: "UCloud: ucloud_disks"
sidebar_current: "docs-ucloud-datasource-disks"
description: |-
  Provides a list of Disk resources in the current region.
---

# ucloud_disks

This data source provides a list of Disk resources according to their Disk ID and disk type.

## Example Usage

```hcl
data "ucloud_disks" "example" {}

output "first" {
  value = data.ucloud_disks.example.disks[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where Disk are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `ids` - (Optional) A list of Disk IDs, all the Disks belong to this region will be retrieved if the ID is `""`.
* `disk_type` - (Optional) The type of disk. Possible values are: `data_disk`as cloud disk, `ssd_data_disk` as SSD cloud disk, `system_disk`as system disk, `ssd_system_disk` as SSD system disk, `rssd_data_disk` as RDMA-SSD cloud disk. 
* `name_regex` - (Optional) A regex string to filter resulting disks by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `disks` - It is a nested type which documented below.
* `total_count` - Total number of Disks that satisfy the condition.

- - -

The attribute (`disks`) support the following:

* `availability_zone` - Availability zone where disk is located.
* `id` - The ID of Disk.
* `name` - The name of Disk.
* `disk_size` - The size of disk. Purchase the size of disk in GB.
* `disk_type` - The type of disk.
* `charge_type` - The charge type of disk. Possible values are: `year` as pay by year, `month` as pay by month, `dynamic` as pay by hour.
* `tag` - A tag assigned to Disk.
* `create_time` - The creation time of Disk, formatted in RFC3339 time string.
* `expire_time` - The expiration time of disk, formatted in RFC3339 time string.
* `status` - The status of disk. Possible values are: `Available`, `InUse`, `Detaching`, `Initializating`, `Failed`, `Cloning`, `Restoring`, `RestoreFailed`.