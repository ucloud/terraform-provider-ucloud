---
layout: "ucloud"
page_title: "UCloud: ucloud_projects"
sidebar_current: "docs-ucloud-datasource-projects"
description: |-
  Provides a list of projects owned by the user.
---

# ucloud_projects

This data source providers a list projects owned by the user according to whether or not be finance account.

## Example Usage

```hcl
data "ucloud_projects" "example" {
    is_finance = "No"
}

output "first" {
    value = "${data.ucloud_instances.example.projects.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `is_finance` - (Optional) To identify if the current account is granted with financial permission, possible values are: "Yes" and "No".
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `projects` - projects is a nested type. projects documented below.
* `total_count` - Total number of project that satisfy the condition.

The attribute (`projects`) support the following:

* `create_time` - The time of creation for instance.
* `id` - The ID of project defined.
* `member_count` - The number of members belongs to the defined project.
* `name` - The name of the defined project.
* `parent_id` - The ID of the parent Project where the sub project belongs to.
* `parent_name` - The name of the parent Project where the sub project belongs to.
* `resource_count` - The number of the resounce instance belong to the defined project.