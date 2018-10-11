---
layout: "ucloud"
page_title: "UCloud: ucloud_instances"
sidebar_current: "docs-ucloud-datasource-instances"
description: |-
  Provides a list of UHost instance resources in the current region.
---

# ucloud_instances

This data source providers a list UHost instance resources according to their availability zone, instance ID and tag.

## Example Usage

```hcl
data "ucloud_instances" "example" {
    availability_zone = "cn-sh2-02"
}

output "first" {
    value = "${data.ucloud_instances.example.instances.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where instances are located. such as: "cn-bj-01". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `ids` - (Optional) The group of IDs of instances that require to be retrieved, all the instances belongs to the defined region will be retrieved if this argument is "".
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `tags` - (Optional) A mapping of tags to assign to instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instances` - instances is a nested type. instances documented below.
* `total_count` - Total number of instance that satisfy the condition.

The attribute (`instances`) support the following:

* `auto_renew` - To identify if the auto renewal is on, possible values are : "Yes" and “No”.
* `cpu` - The number of cores of virtual CPU, measured in "core".
* `memory` - The size of memory, measured in MB.
* `create_time` - The time of creation for EIP.
* `expire_time` - The expiration time for instance.
* `data_disk_category` - The type of disk, the defination of disk type for both system disk and data disk. Possible values are: "LocalDisk" and "Disk" as cloud disk. The "Disk" is not supported in all regions as disk type, please proceed to UCloud console for more details.
* `id` - The ID of instance.
* `instance_charge_type` - The charge type of instance, possible values are: "Year", "Month" and "Dynamic" as pay by hour.
* `name` - The name of the instance.
* `remark` - The remarks of instance.
* `status` - Instance current status. Possible values are "Initializing", "starting", "Running", "Stopping", "Stopped", "Install Fail" and "Rebooting".
* `tag` - A mapping of tags to assign to the instance.