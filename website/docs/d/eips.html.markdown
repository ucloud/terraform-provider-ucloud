---
layout: "ucloud"
page_title: "UCloud: ucloud_eips"
sidebar_current: "docs-ucloud-datasource-eips"
description: |-
  Provides a list of EIP resources in the current region.
---

# ucloud_eips

This data source provides a list of EIP resources (Elastic IP address) according to their EIP ID.

## Example Usage

```hcl
data "ucloud_eips" "example" {}

output "first" {
  value = data.ucloud_eips.example.eips[0].ip_set[0].ip
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional)  A list of Elastic IP IDs, all the EIPs belong to this region will be retrieved if the ID is `""`.
* `name_regex` - (Optional) A regex string to filter resulting eips by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `eips` - It is a nested type which documented below.
* `total_count` - Total number of Elastic IPs that satisfy the condition.

The attribute (`eips`) support the following:

* `bandwidth` - Maximum bandwidth to the elastic public network, measured in Mbps.
* `ip_set` - It is a nested type which documented below.
* `create_time` - The creation time of Elastic IP, formatted in RFC3339 time string.
* `expire_time` - The expiration time for Elastic IP, formatted in RFC3339 time string.
* `charge_mode` - The charge mode of Elastic IP. Possible values are: `traffic` as pay by traffic, `bandwidth` as pay by bandwidth.
* `charge_type` - The charge type of Elastic IP. Possible values are: `year` as pay by year, `month` as pay by month, `dynamic` as pay by hour.
* `name` - The name of Elastic IP.
* `remark` - The remarks of Elastic IP.
* `status` - Elastic IP status. Possible values are: `used` as in use, `free` as available and `freeze` as associating.
* `tag` - A tag assigned to Elastic IP.

The attribute (`ip_set`) support the following:

* `internet_type` - Type of Elastic IP routes.
* `ip` - Elastic IP address.