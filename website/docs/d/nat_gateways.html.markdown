---
subcategory: "VPC"
layout: "ucloud"
page_title: "UCloud: ucloud_nat_gateways"
description: |-
  Provides a list of Nat Gateway resources in the current region.
---

# ucloud_nat_gateways

This data source providers a list of Nat Gateway resources according to their ID and name.

## Example Usage

```hcl
data "ucloud_nat_gateways" "example" {
}

output "first" {
  value = data.ucloud_nat_gateways.example.nat_gateways[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of Nat Gateway IDs, all the Nat Gateways belongs to the defined region will be retrieved if this argument is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting Nat Gateways by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `nat_gateways` - It is a nested type. Nat Gateways documented below.
* `total_count` - Total number of Nat Gateways that satisfy the condition.

- - -

The attribute (`nat_gateways`) support the following:

* `id` - The ID of Nat Gateway.
* `name` - The name of the Nat Gateway.
* `remark` - The remarks of Nat Gateway.
* `tag` - A tag assigned to the Nat Gateway.
* `vpc_id` - The ID of VPC linked to the Nat Gateway.
* `subnet_ids` - The list of subnet ID under the VPC.
* `security_group` -The ID of the associated security group.
* `create_time` - The time of creation for Nat Gateway, formatted in RFC3339 time string.
* `ip_set` - It is a nested type which documented below.

The attribute (`ip_set`) supports the following:

* `internet_type` - Type of Elastic IP routes.
* `ip` - Elastic IP address.