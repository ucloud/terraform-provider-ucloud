---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_udb_mysql_instances"
description: |-
  Provides a list of MySQL instance resources.
---

# ucloud_udb_mysql_instances

This data source provides a list of MySQL instance resources.

## Example Usage

```hcl
data "ucloud_udb_mysql_instances" "example" {}

output "first" {
  value = data.ucloud_udb_mysql_instances.example.udb_mysql_instances[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where the MySQL instances are located. Such as: "cn-bj2-05".
* `ids` - (Optional) A list of MySQL instance IDs.
* `name_regex` - (Optional) A regex string to filter resulting instances by name.
* `vpc_id` - (Optional) The ID of the VPC linked to the instances.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `udb_mysql_instances` - It is a nested type which documented below.
* `total_count` - Total number of instances that satisfy the condition.

- - -

The attribute (`udb_mysql_instances`) support the following:

* `id` - The ID of the instance.
* `name` - The name of the instance.
* `availability_zone` - Availability zone where the instance is located.
* `status` - The current status of the instance.
* `private_ip` - The private IP address assigned to the instance.
* `port` - The port on which the database accepts connections.
* `db_version` - The database version.
* `instance_mode` - The instance mode.
* `memory` - The size of memory in gigabytes (GB).
* `disk_space` - The allocated storage size in gigabytes (GB).
* `cpu` - The number of CPU cores.
* `vpc_id` - The ID of the VPC linked to the instance.
* `subnet_id` - The ID of the subnet linked to the instance.
* `charge_type` - The charge type of the instance.
* `create_time` - The time when the instance was created.
* `expire_time` - The expiration time of the instance.
