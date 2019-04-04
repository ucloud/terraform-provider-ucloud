---
layout: "ucloud"
page_title: "UCloud: ucloud_vpcs"
sidebar_current: "docs-ucloud-datasource-vpcs"
description: |-
  Provides a list of VPC resources in the current region.
---

# ucloud_vpcs

This data source provides a list of VPC resources according to their VPC ID, name.

## Example Usage

```hcl
data "ucloud_vpcs" "example" {
}

output "first" {
    value = "${data.ucloud_vpcs.example.vpcs.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of VPC IDs, all the VPC resources belong to this region will be retrieved if the ID is `""`.
* `name_regex` - (Optional) A regex string to filter resulting VPC resources by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpcs` - It is a nested type which documented below.
* `total_count` - Total number of VPC resources that satisfy the condition.

The attribute (`vpcs`) support the following:

<!-- * `id` - The ID of VPC.
* `name` - The name of VPC.
* `cidr_block` - The cidr block of the desired VPC.
* `type` - The type of VPC.
* `vpc_id` - The id of the VPC that the desired VPC belongs to.
* `create_time` - The time of creation of VPC, formatted in RFC3339 time string.
* `remark` - The remark of the VPC.
* `tag` - A tag assigned to VPC. -->