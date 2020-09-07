---
subcategory: "ULB"
layout: "ucloud"
page_title: "UCloud: ucloud_lb_ssls"
description: |-
  Provides a list of Load Balancer SSL certificate resources.
---

# ucloud_lb_ssls

This data source provides a list of Load Balancer SSL certificate resources according to their Load Balancer SSL certificate resource ID and name.

## Example Usage

```hcl
data "ucloud_lb_ssls" "example" {
}

output "first" {
  value = data.ucloud_lb_ssls.example.lb_ssls[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of LB SSL certificate resource IDs, all the LB SSL certificate resources in the current region will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting LB SSL by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `lb_ssls` - It is a nested type which documented below.
* `total_count` - Total number of LB SSL certificate resources that satisfy the condition.

- - -

The attribute (`lb_ssls`) support the following:

* `id` - The ID of LB SSL certificate resource.
* `name` - The name of LB SSL certificate resource.
* `create_time` - The time of creation for lb ssl, formatted in RFC3339 time string.