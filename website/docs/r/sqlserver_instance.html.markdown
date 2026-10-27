---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_sqlserver_instance"
description: |-
  Provides a SQL Server instance resource.
---

# ucloud_sqlserver_instance

Provides a SQL Server instance resource on the nvme machine type.

## Example Usage

```hcl
resource "ucloud_sqlserver_instance" "example" {
  availability_zone = "cn-bj2-05"
  name              = "tf-example-sqlserver"
  db_version        = "sqlserver-2019"
  cpu               = 2
  memory            = 4000
  disk_space        = 50
  password          = "2018_UClou"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where the SQL Server instance is located. Such as: "cn-bj2-05". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `cpu` - (Required, ForceNew) The number of CPU cores.
* `memory` - (Required, ForceNew) The size of memory in megabytes (MB).
* `disk_space` - (Required, ForceNew) The allocated storage size in gigabytes (GB), range from 20 to 32000.
* `db_version` - (Optional, ForceNew) The SQL Server version, possible values are: `sqlserver-2017`, `sqlserver-2019`, `sqlserver-2022`. (Default: `sqlserver-2019`).
* `name` - (Optional, ForceNew) The name of the SQL Server instance, which contains 6-63 characters. If not specified, terraform will auto-generate a name beginning with `tf-sqlserver-instance`.
* `password` - (Optional, ForceNew) The password for the SQL Server instance which should have 8-30 characters. If not specified, terraform will auto-generate a password.
* `port` - (Optional, ForceNew) The port on which the database accepts connections. (Default: `1433`).
* `charge_type` - (Optional, ForceNew) The charge type of the instance, possible values are: `Month`, `Year`, `Dynamic`. (Default: `Month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance. (Default: `1`).
* `vpc_id` - (Optional, ForceNew) The ID of the VPC linked to the instance.
* `subnet_id` - (Optional, ForceNew) The ID of the subnet linked to the instance.
* `tag` - (Optional, ForceNew) A tag assigned to the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the SQL Server instance.
* `private_ip` - The private IP address assigned to the instance.
* `status` - The current status of the instance.
* `create_time` - The time when the instance was created.
* `expire_time` - The expiration time of the instance.
* `modify_time` - The last modified time of the instance.

## Import

SQL Server instances can be imported using their ID:

```sh
terraform import ucloud_sqlserver_instance.example <instance-id>
```
