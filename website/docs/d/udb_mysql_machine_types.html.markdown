---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_udb_mysql_machine_types"
description: |-
  Provides a list of MySQL machine types.
---

# ucloud_udb_mysql_machine_types

This data source provides a list of MySQL machine types.

## Example Usage

```hcl
data "ucloud_udb_mysql_machine_types" "example" {
  availability_zone = "cn-bj2-05"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) Availability zone where the machine types are available.
* `instance_mode` - (Optional) The instance mode to filter, possible values are: `Normal`, `HA`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `udb_mysql_machine_types` - It is a nested type which documented below.
* `total_count` - Total number of machine types that satisfy the condition.

- - -

The attribute (`udb_mysql_machine_types`) support the following:

* `id` - The ID of the machine type, e.g. `o.mysql2m.small`.
* `cpu` - The number of CPU cores.
* `memory` - The size of memory in gigabytes (GB).
* `description` - The description of the machine type.
* `group` - The memory/cpu ratio group.
* `specification_class` - The specification class, `O` for nvme.
* `storage_class` - The storage class, `CLOUD_RSSD` for nvme.
