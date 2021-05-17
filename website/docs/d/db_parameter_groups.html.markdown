---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_db_parameter_groups"
description: |-
  Provides a list of db parameter group resources in the current region.
---

# ucloud_db_parameter_groups

This data source provides a list of parameter group resources according to their name and availability zone.

## Example Usage

```hcl
data "ucloud_db_parameter_groups" "example" {}

output "first" {
  value = data.ucloud_db_parameter_groups.example.parameter_groups[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where parameter groups are located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `multi_az` - (Optional) Specifies if the replication instance is a multi-az deployment. You cannot set the `availability_zone` parameter if the `multi_az` parameter is set to `true`.
* `name_regex` - (Optional) A regex string to filter resulting parameter groups by name.
* `class_type` - (Optional) The type of the DB instance, Possible values are: `sql` for mysql or percona, `postgresql` for postgresql.  
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `parameter_groups` - It is a nested type which documented below.
* `total_count` - Total number of parameter groups that satisfy the condition.

- - -

The attribute (`parameter_groups`) support the following:

* `id` - The ID of parameter group.
* `name` - The name of parameter group.
* `availability_zone` - Availability zone where parameter group is located.
* `engine` - The type of database instance engine used by the parameter group.
* `engine_version` - The database instance engine version used by the parameter group.