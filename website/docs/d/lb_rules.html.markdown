---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_rules"
sidebar_current: "docs-ucloud-datasource-lb-rules"
description: |-
  Provides a list of Load Balancer Rule resources belong to the Load Balancer listener.
---

# ucloud_lb_rules

This data source provides a list of Load Balancer Rule resources according to their Load Balancer Rule ID.

## Example Usage

```hcl
data "ucloud_lb_rules" "example" {
  load_balancer_id = "ulb-xxx"
  listener_id      = "vserver-xxx"
}

output "first" {
  value = data.ucloud_lb_rules.example.lb_rules[0].id
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of a load balancer.
* `listener_id` - (Required) The ID of a listener server.

- - -

* `ids` - (Optional) A list of LB Rule IDs, all the LB Rules belong to the Load Balancer listener will be retrieved if the ID is `[]`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `lb_rules` - It is a nested type which documented below.
* `total_count` - Total number of LB Rules that satisfy the condition.

- - -

The attribute (`lb_rules`) support the following:

* `id` - The ID of LB Rule.
* `path` - (Optional) The path of Content forward matching fields. `path` and `domain` cannot coexist.
* `domain` - (Optional) The domain of content forward matching fields. `path` and `domain` cannot coexist.