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
    availability_zone = "cn-sh2-02"
    name              = "tf-example-disk"
    disk_size         = 10
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) The Zone to create the disk in.
* `name` - (Required)  The name of disk, should have 6 - 63 characters and only support chinese, english, numbers, '-', '_'.
* `disk_size` - (Required) Purchase the size of disk. Volume is from 1 to 8000GB as cloud disk, from 1 to 4000GB as ssd cloud disk.
* `disk_type` - (Optional)the type of disk. Possible values are: "DataDisk" as cloud disk, "SSDDataDisk" as ssd cloud disk, the default is "DataDisk".
* `disk_charge_type` - (Optional) Charge type of disk. Possible values are: "Year" as pay by year, "Month" as pay by month, "Dynamic" as pay by hour. The default value is "Dynamic".
* `disk_duration` - (Optional) The duration that you will buy the resource, the default value is "1". It is not required when "Dynamic" (pay by hour), the value is "0" when pay by month and the instance will be vaild till the last day of that month.
* `tag` - (Optional) A mapping of tags to assign to the disk, the default value is"Default"(means no tag assigned).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `create_time` - The time of creation for disk.
* `expire_time` - The expiration time for disk.
* `status` -  status. Possible values are: "Available", "InUse", "Detaching", "Initializating", "Failed", "Cloning", "Restoring", "RestoreFailed".