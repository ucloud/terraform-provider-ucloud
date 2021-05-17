---
subcategory: "UFS"
layout: "ucloud"
page_title: "UCloud: ucloud_ufs_volumes"
description: |-
  Provides a list of UFS Volume resources in the current region.
---

# ucloud_ufs_volumes

This data source provides a list of UFS Volume resources according to their UFS Volume ID and ufs volume name.

## Example Usage

```hcl
data "ucloud_ufs_volumes" "example" {}

output "first" {
  value = data.ucloud_ufs_volumes.example.ufs_volumes[0].id
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of UFS Volume IDs, all the UFS Volumes belong to this region will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting UFS Volumes by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ufs_volumes` - It is a nested type which documented below.
* `total_count` - Total number of UFS Volumes that satisfy the condition.

- - -

The attribute (`ufs_volumes`) support the following:

* `id` - The ID of UFS Volume.
* `name` - The name of UFS Volume.
* `tag` - A tag assigned to UFS Volume.
* `remark` - A remark assigned to UFS Volume.  
* `size` - The size of ufs volume. Purchase the size of ufs volume in GB.
* `storage_type` - The storage type of ufs volume.
* `protocol_type` - The protocol type of ufs volume.
* `create_time` - The creation time of UFS Volume, formatted in RFC3339 time string.
* `expire_time` - The expiration time of ufs volume, formatted in RFC3339 time string.