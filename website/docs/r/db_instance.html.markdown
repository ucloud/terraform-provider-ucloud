---
layout: "ucloud"
page_title: "UCloud: ucloud_db_instance"
sidebar_current: "docs-ucloud-resource-db-instance"
description: |-
  Provides a Database instance resource.
---

# ucloud_db_instance

Provides a Database instance resource.

~> **Note**  Please do confirm if any task pending submission before reset your password, since the password reset will take effect immediately.
## Example Usagek

```hcl
# Query availability zone
data "ucloud_zones" "default" {
}

# Create database instance
resource "ucloud_db_instance" "master" {
  name              = "tf-example-db"
  instance_storage  = 20
  instance_type     = "mysql-ha-1"
  engine            = "mysql"
  engine_version    = "5.7"
  password          = "2018_dbInstance"
}
```
## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) Availability zone where database instance is located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `engine` - (Required) The type of database engine, possible values are: "mysql", "percona".
* `engine_version` - (Required) The database engine version, possible values are: "5.5", "5.6", "5.7".
    - 5.5/5.6/5.7 for mysql and percona engine. 
* `name` - (Optional) The name of database instance, which contains 6-63 characters and only support Chinese, English, numbers, '-', '_', '.', ',', '[', ']', ':'. If not specified, terraform will auto-generate a name beginning with `tf-db-instance`.
* `instance_storage` - (Required) Specifies the allocated storage size in gigabytes (GB), range from 20 to 4500GB. The volume adjustment must be a multiple of 10 GB. The maximum disk volume for SSD type are：
    - 500GB if the memory chosen is equal or less than 6GB;
    - 1000GB if the memory chosen is from 8 to 16GB;
    - 2000GB if the memory chosen is 24GB or 32GB;
    - 3500GB if the memory chosen is 48GB or 64GB;
    - 4500GB if the memory chosen is equal or more than 96GB;
* `instance_type` - (Required) The type of database instance, please visit the [instance type table](https://www.terraform.io/docs/providers/ucloud/appendix/db_instance_type.html).

- - -

* `standby_zone` - (Optional) Availability zone where the standby database instance is located for the high availability database instance with multiple zone; The disaster recovery of data center can be activated by switching to the standby database instance for the high availability database instance.
* `password` - (Optional) The password for the database instance which should have 8-30 characters. It must contain at least 3 items of Capital letters, small letter, numbers and special characters. The special characters include `-_`. If not specified, terraform will auto-generate a password.
* `port` - (Optional) The port on which the database accepts connections, the default port is 3306 for mysql and percona.
* `charge_type` - (Optional) The charge type of db instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional) The duration that you will buy the db instance (Default: `1`). The value is `0` when pay by month and the instance will be vaild till the last day of that month. It is not required when `dynamic` (pay by hour).
* `vpc_id` - (Optional) The ID of VPC linked to the database instances.
* `subnet_id` - (Optional) The ID of subnet.
* `backup_count` - (Optional) Specifies the number of backup saved per week, it is 7 backups saved per week by default.
* `backup_begin_time` - (Optional) Specifies when the backup starts, measured in hour, it starts at one o'clock of 1, 2, 3, 4 in the morning by default.
* `backup_date` - (Optional) Specifies whether the backup took place from Sunday to Saturday by displaying 7 digits. 0 stands for backup disabled and 1 stands for backup enabled. The rightmost digit specifies whether the backup took place on Sunday, and the digits from right to left specify whether the backup took place from Monday to Saturday, it's mandatory required to backup twice per week at least. such as: digits "1100000" stands for the backup took place on Saturday and Friday.
* `backup_black_list` - (Optional) The backup for database such as "test.%" or table such as "city.address" specified in the black lists are not supported.
* `tag` - (Optional) A tag assigned to database instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `status` - Specifies the status of database, possible values are: `Init`, `Fail`, `Starting`, `Running`, `Shutdown`, `Shutoff`, `Delete`, `Upgrading`, `Promoting`, `Recovering` and `Recover fail`.
* `private_ip` - The private IP address assigned to the database instance.
* `create_time` - The creation time of database, formatted by RFC3339 time string.
* `expire_time` - The expiration time of database, formatted by RFC3339 time string.
* `modify_time` - The modification time of database, formatted by RFC3339 time string.

## Import

DB Instance can be imported using the `id`, e.g.

```
$ terraform import ucloud_db_instance.example udbha-abc123456
```