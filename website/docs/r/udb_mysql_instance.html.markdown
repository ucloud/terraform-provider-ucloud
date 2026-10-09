---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_udb_mysql_instance"
description: |-
  Provides a MySQL instance resource.
---

# ucloud_udb_mysql_instance

Provides a MySQL instance resource on the nvme machine type.

## Example Usage

```hcl
resource "ucloud_udb_mysql_instance" "example" {
  availability_zone = "cn-bj2-05"
  name              = "tf-example-mysql"
  db_version        = "mysql-8.0"
  machine_type      = "o.mysql2m.small"
  disk_space        = 20
  password          = "2018_UClou"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where the MySQL instance is located. Such as: "cn-bj2-05". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `db_version` - (Optional, ForceNew) The MySQL database version, possible values are: `mysql-5.7`, `mysql-8.0`, `mysql-8.4`. (Default: `mysql-8.0`).
* `machine_type` - (Required) The machine type of the MySQL instance, e.g. `o.mysql2m.small` for 1C2G. See `ucloud_udb_mysql_machine_types` data source for the full list. Changing this triggers a resize (CPU/memory upgrade).
* `name` - (Optional, ForceNew) The name of the MySQL instance, which contains 6-63 characters. If not specified, terraform will auto-generate a name beginning with `tf-mysql-instance`.
* `password` - (Optional, ForceNew) The password for the MySQL instance which should have 8-30 characters. If not specified, terraform will auto-generate a password.
* `instance_mode` - (Optional, ForceNew) The instance mode, possible values are: `Normal`, `HA`. (Default: `HA`).
* `disk_space` - (Required) The allocated storage size in gigabytes (GB), range from 20 to 32000. Changing this triggers a resize (storage upgrade).
* `port` - (Optional, ForceNew) The port on which the database accepts connections. (Default: `3306`).
* `charge_type` - (Optional, ForceNew) The charge type of the instance, possible values are: `Month`, `Year`, `Dynamic`. (Default: `Month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance. (Default: `1`).
* `param_group_id` - (Optional, ForceNew) The ID of the parameter group. If not specified, the default template for the db version is used.
* `vpc_id` - (Optional, ForceNew) The ID of the VPC linked to the instance.
* `subnet_id` - (Optional, ForceNew) The ID of the subnet linked to the instance.
* `tag` - (Optional, ForceNew) A tag assigned to the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the MySQL instance.
* `private_ip` - The private IP address assigned to the instance.
* `status` - The current status of the instance.
* `cpu` - The number of CPU cores.
* `memory` - The size of memory in gigabytes (GB).
* `create_time` - The time when the instance was created.
* `expire_time` - The expiration time of the instance.
* `modify_time` - The last modified time of the instance.

## Import

MySQL instances can be imported using their ID:

```sh
terraform import ucloud_udb_mysql_instance.example <instance-id>
```
