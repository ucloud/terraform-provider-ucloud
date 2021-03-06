---
subcategory: "UFS"
layout: "ucloud"
page_title: "UCloud: ucloud_ufs_volume"
description: |-
  Provides a UFS Volume resource.
---

# ucloud_ufs_volume

Provides a UFS Volume resource.

## Example Usage

```hcl
resource "ucloud_ufs_volume" "foo" {
	name  	 	  = "tf-acc-ufs-volume-basic"
	remark 		  = "test"
	tag           = "tf-acc"
	size      	  = 500 
	storage_type  = "Basic"
	protocol_type = "NFSv4"
}
```

## Argument Reference

The following arguments are supported:

* `size` - (Required, ForceNew) The size of the UFS Volume, measured in GB (GigaByte), 500 - 100000 for `Basic` storage type, 100 - 20000 for `Advanced` storage type.
* `storage_type` - (Required, ForceNew) The storage type of the UFS Volume. Possible values are: `Basic`, `Advanced`.
* `protocol_type` - (Required, ForceNew) The protocol_type of the UFS Volume. Possible values are: `NFSv3`, `NFSv4`.

- - -

* `charge_type` - (Optional, ForceNew) The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `name` - (Optional, ForceNew) The name of instance, expected value to be 6 - 63 characters and only support english, numbers, '-', '_', and can not prefix with '-'. If not specified, terraform will auto-generate a name beginning with `tf-ufs-volume-`.
* `tag` - (Optional, ForceNew) A tag assigned to UFS Volume, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `remark` - (Optional, ForceNew) The remarks of instance. (Default: `""`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource UFS Volume.
* `create_time` - The time of creation of UFS Volume, formatted in RFC3339 time string.
* `expire_time` - The expiration time of UFS Volume, formatted in RFC3339 time string.