---
subcategory: "UFS"
layout: "ucloud"
page_title: "UCloud: ucloud_ufss"
description: |-
  Provides a list of UFS resources in the current region.
---

# ucloud_ufss

This data source provides a list of UFS resources according to their UFS ID and ufs name.

## Example Usage

```hcl
data "ucloud_ufss" "example" {}

output "first" {
  value = data.ucloud_ufss.example.ufss[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of UFS IDs, all the UFSs belong to this region will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting UFSs by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ufss` - It is a nested type which documented below.
* `total_count` - Total number of UFSs that satisfy the condition.

- - -

The attribute (`ufss`) support the following:

* `id` - The ID of UFS.
* `name` - The name of UFS.
* `tag` - A tag assigned to UFS.
* `remark` - A remark assigned to UFS.  
* `size` - The size of ufs. Purchase the size of ufs in GB.
* `storage_type` - The storage type of ufs.
* `protocol_type` - The protocol type of ufs.
* `create_time` - The creation time of UFS, formatted in RFC3339 time string.
* `expire_time` - The expiration time of ufs, formatted in RFC3339 time string.