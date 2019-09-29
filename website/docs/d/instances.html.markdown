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
  value = data.ucloud_instances.example.instances[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where instances are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `ids` - (Optional) A list of instance IDs, all the instances belongs to the defined region will be retrieved if this argument is "".
* `name_regex` - (Optional) A regex string to filter resulting instances by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `tag` - (Optional) A tag assigned to instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instances` - It is a nested type. instances documented below.
* `total_count` - Total number of instances that satisfy the condition.

- - -

The attribute (`instances`) support the following:

* `availability_zone` - Availability zone where instances are located.
* `id` - The ID of instance.
* `name` - The name of the instance.
* `cpu` - The number of cores of virtual CPU, measureed in core.
* `memory` - The size of memory, measured in MB (Megabyte).
* `instance_type` - The type of instance.
* `charge_type` - The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour.
* `auto_renew` - Whether to renew an instance automatically or not.
* `remark` - The remarks of instance.
* `tag` - A tag assigned to the instance.
* `status` - Instance current status. Possible values are `Initializing`, `Starting`, `Running`, `Stopping`, `Stopped`, `Install Fail` and `Rebooting`.
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `expire_time` - The expiration time for instance, formatted in RFC3339 time string.
* `private_ip` - The private IP address assigned to the instance.
* `vpc_id` - The ID of VPC linked to the instance.
* `subnet_id` - The ID of subnet linked to the instance.
* `ip_set` - It is a nested type which documented below.
* `disk_set` - It is a nested type which documented below.

The attribute (`disk_set`) supports the following:

* `id` - The ID of disk.
* `size` - The size of disk, measured in GB (Gigabyte).
* `type` - The type of disk.
* `is_boot` - Specifies whether boot disk or not.

The attribute (`ip_set`) supports the following:

* `internet_type` - Type of Elastic IP routes. Possible values are: `International` as internaltional BGP IP, `BGP` as china BGP IP and `Private` as private IP.
* `ip` - Elastic IP address.