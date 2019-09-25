---
layout: "ucloud"
page_title: "UCloud: ucloud_zones"
sidebar_current: "docs-ucloud-datasource-zones"
description: |-
  Provides a list of available zones in the current region.
---

# ucloud_zones

This data source provides a list of available zones in the current region.

## Example Usage

```hcl
data "ucloud_zones" "example" {}

output "first" {
  value = data.ucloud_zones.example.zones[0].id
}
```

## Argument Reference

The following arguments are supported:

* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `zones` - It is a nested type which documented below.
* `total_count` - Total number of zones that satisfy the condition.

- - -

The attribute (`zones`) support the following:

* `id` -  The ID of availability zone.