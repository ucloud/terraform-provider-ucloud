---
subcategory: "US3"
layout: "ucloud"
page_title: "UCloud: ucloud_us3_bucket"
description: |-
  Provides a US3 bucket resource.
---

# ucloud_us3_bucket

Provides a US3 bucket resource.

## Example Usage

```hcl
resource "ucloud_us3_bucket" "foo" {
	name  	= "tf-acc-us3-bucket-basic"
    type    = "private"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, ForceNew) The name of the US3 bucket, expected value to be:
    - 3 - 63 characters.
    - only support lowercase-letters, numbers, '-'.
    - not prefix with '-', 'www', 'cn-bj', 'hk'.
    - not suffix with '-'.
* `type` - (Required) The type of the US3 bucket. Possible values are: `public`, `private`.

- - -

* `tag` - (Optional, ForceNew) A tag assigned to UFS, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource US3 bucket.
* `create_time` - The time of creation of US3 bucket, formatted in RFC3339 time string.
* `source_domain_names` - The list of source domain name.