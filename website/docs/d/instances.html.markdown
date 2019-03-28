---
layout: "ucloud"
page_title: "UCloud: ucloud_instances"
sidebar_current: "docs-ucloud-datasource-instances"
description: |-
  Provides a list of UHost instance resources in the current region.
---

# ucloud_instances

This data source providers a list of UHost instance resources according to their availability zone, instance ID and tag.

## Example Usage

```hcl
data "ucloud_instances" "example" {
    availability_zone = "cn-bj2-02"
}

output "first" {
    value = "${data.ucloud_instances.example.instances.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where instances are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `ids` - (Optional) A list of instance IDs, all the instances belongs to the defined region will be retrieved if this argument is "".
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `tag` - (Optional) A mapping of tags to assign to instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instances` - It is a nested type. instances documented below.
* `total_count` - Total number of instance that satisfy the condition.

The attribute (`instances`) support the following:

* `auto_renew` - To identify if the auto renewal is on, possible values are : "Yes" and “No”.
* `cpu` - The number of cores of virtual CPU, measured in "core".
* `memory` - The size of memory, measured in MB.
* `create_time` - The creation time of instance.
* `expire_time` - The expiration time of instance.
* `id` - The ID of instance.
* `instance_charge_type` - The charge type of instance, possible values are: "Year", "Month" and "Dynamic" as pay by hour.
* `name` - The name of the instance.
* `remark` - The remarks of instance.
* `status` - Instance current status. Possible values are "Initializing", "starting", "Running", "Stopping", "Stopped", "Install Fail" and "Rebooting".
* `tag` - A mapping of tags to assign to the instance.
* `ip_set` - ip_set is a nested type. ip_set documented below.
* `disk_set` - disk_set is a nested type. disk_set documented below.

The attribute (`disk_set`) support the following:

* `disk_id` - The ID of disk.
* `size` - The size of disk，measured in GB (Gigabyte).
* `disk_type` - The type of disk.
* `is_boot` - whether or not boot disk.

The attribute (`ip_set`) support the following:

* `type` - IP type.
* `ip` - IP address.