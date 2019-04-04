---
layout: "ucloud"
page_title: "UCloud: ucloud_disk"
sidebar_current: "docs-ucloud-resource-disk"
description: |-
  Provides a Cloud Disk resource.
---

# ucloud_disk

Provides a Cloud Disk resource.

## Example Usage

```hcl
resource "ucloud_disk" "example" {
    availability_zone = "cn-bj2-02"
    name              = "tf-example-disk"
    disk_size         = 10
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required)  Availability zone where cloud disk is located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist).
* `disk_size` - (Required) The size of disk. Purchase the size of disk in GB. 1-8000 for a cloud disk, 1-4000 for SSD cloud disk.
* `name` - (Optional)  The name of disk, should have 6-63 characters and only support Chinese, English, numbers, '-', '_'. If not specified, terraform will autogenerate a name beginning with `tf-disk`.
* `disk_type` - (Optional) The type of disk. Possible values are: `data_disk`as cloud disk, `ssd_data_disk` as ssd cloud disk. (Default: `data_disk`).
* `charge_type` - (Optional) Charge type of disk. Possible values are: `year` as pay by year, `month` as pay by month, `dynamic` as pay by hour. (Default: `month`).
* `duration` - (Optional) The duration that you will buy the resource. (Default: `1`). It is not required when `dynamic` (pay by hour), the value is `0` when `month`(pay by month) and the disk will be vaild till the last day of that month.
* `tag` - (Optional) A tag assigned to VPC, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of disk, formatted in RFC3339 time string.
* `expire_time` - The expiration time of disk, formatted in RFC3339 time string.
* `status` -  The status of disk. Possible values are: `Available`, `InUse`, `Detaching`, `Initializating`, `Failed`, `Cloning`, `Restoring`, `RestoreFailed`.
