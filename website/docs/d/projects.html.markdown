---
layout: "ucloud"
page_title: "UCloud: ucloud_projects"
sidebar_current: "docs-ucloud-datasource-projects"
description: |-
  Provides a list of projects owned by the user.
---

# ucloud_projects

This data source providers a list of projects owned by user with finance permission.

## Example Usage

```hcl
data "ucloud_projects" "example" {
    is_finance = false
}

output "first" {
    value = "${data.ucloud_instances.example.projects.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `is_finance` - (Optional) To identify if the current account is granted with financial permission.
* `name_regex` - (Optional) A regex string to filter resulting projects by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `projects` - It is a nested type which documented below.
* `total_count` - Total number of projects that satisfy the condition.

The attribute (`projects`) support the following:

* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `id` - The ID of project defined.
* `member_count` - The number of members belongs to the defined project.
* `name` - The name of the defined project.
* `parent_id` - The ID of the parent project where the sub project belongs to.
* `parent_name` - The name of the parent project where the sub project belongs to.
* `resource_count` - The number of the resounce instance belong/s to the defined project.