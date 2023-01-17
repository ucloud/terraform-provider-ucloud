---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_db_backups"
description: |-
  Provides a list of database backups in the current region.
---

# ucloud_db_backups

This data source provides a list of database backups according to their
name, availability zone and project.

## Example Usage

```hcl

data "ucloud_db_backups" "example" {
  availability_zone = "cn-bj2-05"
  name_regex        = "init.*"
}

output "backups" {
  value = data.ucloud_db_backups.example.db_backups
}

```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where database instances are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `project_id` - (Optional) id of the projects to which the backups belong
* `name_regex` - (Optional) A regex string to filter backups by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `db_backups` - It is a nested type which documented below.
* `total_count` - Total number of database backups that satisfy the condition.

- - -

The attribute (`db_backups`) support the following:

* `backup_id` - id of the database backup which can used to seed new database instance
* `backup_name` - name of the database backup
* `backup_size` - size of the database backup in bytes
* `backup_time` - time when the backup was created
* `backup_end_time` - time when the backup was completed
* `backup_type` - type of backup, 0: auto-backup, 1: manual-backup
* `db_id` - id of the database instance
* `db_name` - name of the database instance
* `state` - State of backup
* `backup_zone` - availability zone where the backup database locates, empty
  for non-master-slave instances
* `zone` - availability zone where the database instance locates
