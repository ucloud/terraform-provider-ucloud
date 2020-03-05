---
layout: "ucloud"
page_title: "UCloud: ucloud_db_instances"
sidebar_current: "docs-ucloud-datasource-db-instances"
description: |-
  Provides a list of database instance resources in the current region.
---

# ucloud_db_instances

This data source provides a list of database instance resources according to their database instance ID and name.

## Example Usage

```hcl
data "ucloud_db_instances" "example" {}

output "first" {
  value = data.ucloud_db_instances.example.db_instances[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where database instances are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `ids` - (Optional) A list of database instance IDs, all the database instances belong to this region will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting database instances by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `db_instances` - It is a nested type which documented below.
* `total_count` - Total number of database instances that satisfy the condition.

- - -

The attribute (`db_instances`) support the following:

* `availability_zone` - Availability zone where database instance is located.
* `id` - The ID of database instance.
* `name` - The name of database instance.
* `instance_type` - Specifies the type of database instance.
* `standby_zone` - Availability zone where the standby database instance is located for the high availability database instance with multiple zone.
* `vpc_id` - The ID of VPC linked to the database instances.
* `subnet_id` - The ID of subnet linked to the database instances.
* `engine` - The type of database instance engine.
* `engine_version` - The database instance engine version.
* `port` - The port on which the database instance accepts connections.
* `private_ip` - The private IP address assigned to the database instance.
* `instance_storage` - Specifies the allocated storage size in gigabytes (GB).
* `charge_type` - The charge type of db instance.
* `backup_count` - Specifies the number of backup saved per week.
* `backup_begin_time` - Specifies when the backup starts, measured in hour.
* `backup_date` - Specifies whether the backup took place from Sunday to Saturday by displaying 7 digits. 0 stands for backup disabled and 1 stands for backup enabled. The rightmost digit specifies whether the backup took place on Sunday, and the digits from right to left specify whether the backup took place from Monday to Saturday, it's mandatory required to backup twice per week at least. such as: digits "1100000" stands for the backup took place on Saturday and Friday.
* `backup_black_list` - The backup for database instance such as "test.%" or table such as "city.address" specified in the black lists are not supported.
* `tag` - A tag assigned to database instance.
* `status` - Specifies the status of database instance , possible values are: `Init`, `Fail`, `Starting`, `Running`, `Shutdown`, `Shutoff`, `Delete`, `Upgrading`, `Promoting`, `Recovering` and `Recover fail`.
* `create_time` - The creation time of database instance , formatted by RFC3339 time string.
* `expire_time` - The expiration time of database instance , formatted by RFC3339 time string.
* `modify_time` - The modification time of database instance , formatted by RFC3339 time string.