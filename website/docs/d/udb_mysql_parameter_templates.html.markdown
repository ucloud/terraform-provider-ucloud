---
subcategory: "UDB"
layout: "ucloud"
page_title: "UCloud: ucloud_udb_mysql_parameter_templates"
description: |-
  Provides a list of MySQL parameter templates.
---

# ucloud_udb_mysql_parameter_templates

This data source provides a list of MySQL parameter templates.

## Example Usage

```hcl
data "ucloud_udb_mysql_parameter_templates" "example" {
  db_version = "mysql-8.0"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where the parameter templates are located.
* `db_version` - (Optional) The database version to filter, e.g. `mysql-8.0`.
* `name_regex` - (Optional) A regex string to filter resulting templates by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `udb_mysql_parameter_templates` - It is a nested type which documented below.
* `total_count` - Total number of templates that satisfy the condition.

- - -

The attribute (`udb_mysql_parameter_templates`) support the following:

* `id` - The ID of the parameter template.
* `name` - The name of the parameter template.
