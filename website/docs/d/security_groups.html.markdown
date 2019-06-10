---
layout: "ucloud"
page_title: "UCloud: ucloud_security_groups"
sidebar_current: "docs-ucloud-datasource-security-groups"
description: |-
  Provides a list of Security Group resources in the current region.
---

# ucloud_security_groups

This data source provides a list of Security Group resources according to their Security Group ID, name and resource id.

## Example Usage

```hcl
data "ucloud_security_groups" "example" {}

output "first" {
    value = data.ucloud_security_groups.example.security_groups[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of Security Group IDs, all the Security Group resources belong to this region will be retrieved if the ID is `""`.
* `name_regex` - (Optional) A regex string to filter resulting Security Group resources by name.
* `type` - (Optional) The type of Security Group. Possible values are: `recommend_web` as the default Web security group that UCloud recommend to users, default opened port include 80, 443, 22, 3389, `recommend_non_web` as the default non Web security group that UCloud recommend to users, default opened port include 22, 3389, `user_defined` as the security groups defined by users. You may refer to [security group](https://docs.ucloud.cn/network/firewall/firewall.html).
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `security_groups` - It is a nested type which documented below.
* `total_count` - Total number of Security Group resources that satisfy the condition.

The attribute (`security_groups`) support the following:

* `id` - The ID of Security Group.
* `name` - The name of Security Group.
* `rules` - It is a nested type which documented below.
* `type` - The type of Security Group.
* `remark` - The remarks of the security group.
* `tag` - A tag assigned to the security group.
* `create_time` - The time of creation for the security group, formatted in RFC3339 time string.

The attribute (`rules`) support the following:

* `cidr_block` - The cidr block of source.
* `policy` - Authorization policy. Can be either `accept` or `drop`.
* `port_range` - The range of port numbers, range: 1-65535. (eg: `port` or `port1-port2`).
* `priority` - Rule priority. Can be `high`, `medium`, `low`.
* `protocol` - The protocol. Can be `tcp`, `udp`, `icmp`, `gre`.
