---
layout: "ucloud"
page_title: "UCloud: ucloud_subnets"
sidebar_current: "docs-ucloud-datasource-subnets"
description: |-
  Provides a list of Subnet resources in the current region.
---

# ucloud_subnets

This data source provides a list of Subnet resources according to their Subnet ID, name and the VPC they belong to.

## Example Usage

```hcl
data "ucloud_subnets" "example" {
  vpc_id = "uvnet-xxx"
}

output "first" {
  value = data.ucloud_subnets.example.subnets[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of Subnet IDs, all the Subnet resources belong to this region will be retrieved if the ID is `[]`.
* `vpc_id` - (Optional) The id of the VPC that the desired Subnet belongs to.
* `name_regex` - (Optional) A regex string to filter resulting Subnet resources by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `subnets` - It is a nested type which documented below.
* `total_count` - Total number of Subnet resources that satisfy the condition.

- - -

The attribute (`subnets`) support the following:

* `id` - The ID of Subnet.
* `name` - The name of Subnet.
* `cidr_block` - The cidr block of the desired Subnet.
* `create_time` - The time of creation of Subnet, formatted in RFC3339 time string.
* `remark` - The remark of the Subnet.
* `tag` - A tag assigned to Subnet.