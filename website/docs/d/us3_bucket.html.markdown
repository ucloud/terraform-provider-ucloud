---
subcategory: "US3"
layout: "ucloud"
page_title: "UCloud: ucloud_us3_buckets"
description: |-
  Provides a list of US3 Bucket resources in the current region.
---

# ucloud_us3_buckets

This data source provides a list of US3 Bucket resources according to us3 bucket name.

## Example Usage

```hcl
data "ucloud_us3_buckets" "example" {}

output "first" {
  value = data.ucloud_us3_buckets.example.us3_buckets[0].name
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) A regex string to filter resulting US3 Buckets by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `us3_buckets` - It is a nested type which documented below.
* `total_count` - Total number of US3 Buckets that satisfy the condition.

- - -

The attribute (`us3_buckets`) support the following:

* `name` - The name of US3 Bucket.
* `tag` - A tag assigned to US3 Bucket.
* `type` - (Required) The type of the US3 bucket.
* `create_time` - The creation time of US3 Bucket, formatted in RFC3339 time string.
* `src_domain_names` - The list of src domain name.