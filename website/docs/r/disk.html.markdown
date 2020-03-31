---
layout: "ucloud"
page_title: "UCloud: ucloud_disk"
sidebar_current: "docs-ucloud-resource-disk"
description: |-
  Provides a Cloud Disk resource.
---

# ucloud_disk

Provides a Cloud Disk resource.

~> **Note** If the disk have attached to the instance, the instance will reboot automatically to make the change take effect when update the  `disk_size`.

## Example Usage

```hcl
# Query availability zone
data "ucloud_zones" "default" {}

# Create cloud disk
resource "ucloud_disk" "example" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-example-disk"
  disk_size         = 10
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew)  Availability zone where cloud disk is located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist).
* `disk_size` - (Required) The size of disk. Purchase the size of disk in GB. 1-8000 for a cloud disk, 1-4000 for SSD cloud disk. If the disk have attached to the instance, the instance will reboot automatically to make the change take effect when update the  `disk_size`.

- - -

* `name` - (Optional)  The name of disk, should have 6-63 characters and only support Chinese, English, numbers, '-', '_'. If not specified, terraform will auto-generate a name beginning with `tf-disk`.
* `disk_type` - (Optional, ForceNew) The type of disk. Possible values are: `data_disk`as cloud disk, `ssd_data_disk` as ssd cloud disk, `rssd_data_disk` as RDMA-SSD cloud disk (the `rssd_data_disk` only be supported in `cn-bj2-05`).(Default: `data_disk`).
* `charge_type` - (Optional, ForceNew) Charge type of disk. Possible values are: `year` as pay by year, `month` as pay by month, `dynamic` as pay by hour. (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the resource. (Default: `1`). It is not required when `dynamic` (pay by hour), the value is `0` when `month`(pay by month) and the disk will be vaild till the last day of that month.
* `tag` - (Optional, ForceNew) A tag assigned to VPC, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation of disk, formatted in RFC3339 time string.
* `expire_time` - The expiration time of disk, formatted in RFC3339 time string.
* `status` -  The status of disk. Possible values are: `Available`, `InUse`, `Detaching`, `Initializating`, `Failed`, `Cloning`, `Restoring`, `RestoreFailed`.

## Import

Disk can be imported using the `id`, e.g.

```
$ terraform import ucloud_disk.example bsm-abcdefg
```