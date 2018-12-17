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
    value = "${data.ucloud_eips.example.eips.0.ip_set.0.ip}"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) The IDs of Elastic IP, all the EIPs belong to this region will be retrieved if the ID is `""`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `eips` - eips is a nested type which documented below.
* `total_count` - Total number of Elastic IP that satisfy the condition.

The attribute (`eips`) support the following:

* `bandwidth` - Maximum bandwidth to the elastic public network, measured in Mbps.
* `ip_set` - It is a nested type which documented below.
* `create_time` - The time of creation for Elastic IP, formatted in RFC3339 time string.
* `expire_time` - The expiration time for Elastic IP, formatted in RFC3339 time string.
* `charge_mode` - Elastic IP charge mode. Possible values are: `traffic` as pay by traffic, `bandwidth` as pay by bandwidth.
* `charge_type` - Elastic IP Charge type. Possible values are: `year` as pay by year, `month` as pay by month, `dynamic` as pay by hour.
* `name` - The name of Elastic IP.
* `remark` - The remarks of Elastic IP.
* `status` - Elastic IP status. Possible values are: `used` as in use, `free` as available and `freeze` as associating.
* `tag` - (Optional) A mapping of tags to assign to VPC, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

The attribute (`ip_set`) support the following:

* `internet_type` - Type of Elastic IP routes.
* `ip` - Elastic IP address