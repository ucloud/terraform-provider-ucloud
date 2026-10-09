---
subcategory: "UHost"
layout: "ucloud"
page_title: "UCloud: ucloud_images"
description: |-
  Provides a list of available image resources in the current region.
---

# ucloud_images

This data source providers a list of available image resources according to their availability zone, image ID and other fields.

## Example Usage

```hcl
data "ucloud_images" "example" {
  availability_zone = "cn-bj2-02"
  image_type        = "base"
  name_regex        = "^CentOS 7.[1-2] 64"
  most_recent       = true
}

output "first" {
  value = data.ucloud_images.example.images[0].id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where images are located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist).
* `name_regex` - (Optional) A regex string to filter resulting images by name. (Such as: `^CentOS 7.[1-2] 64` means CentOS 7.1 of 64-bit operating system or CentOS 7.2 of 64-bit operating system, "^Ubuntu 16.04 64" means Ubuntu 16.04 of 64-bit operating system).
* `image_type` - (Optional) The type of image. Possible values are: `base` as standard image, `business` as owned by market place, and `custom` as custom-image, all the image types will be retrieved by default.
* `os_type` - (Optional) The type of OS. Possible values are: `linux` and `windows`, all the OS types will be retrieved by default.
* `most_recent` - (Optional) If more than one result is returned, use the most recent image.
* `func_type` - (Optional) The category of the image. Possible values are: `gpu` as GPU industry image, `app` as image dedicated to lightweight cloud host, and `uhost` as UHost industry image from marketplace. An invalid value will be ignored by the API.
* `tag` - (Optional) The ID of business group, default is `Default`.
* `include_price` - (Optional) Whether to return the price of industry images, the `price_set` attribute will be empty unless this argument is `true`.
* `image_id` - (Optional) The ID of image.
 ~> **Note** this argument conflicts with `ids`.
* `ids` - (Optional) A list of image IDs, all the images belong to this region will be retrieved if the ID is `[]`.
 ~> **Note** this argument conflicts with `image_id`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `images` - It is a nested type which documented below.
* `total_count` - Total number of images that satisfy the condition.

- - -

The attribute (`images`) support the following:

* `availability_zone` - Availability zone where image is located.
* `create_time` - The time of creation for image, formatted in RFC3339 time string.
* `features` - To identify if any particular feature belongs to the instance, the value is `NetEnhnced` as I/O enhanced instance for now.
* `description` - The description of image if any.
* `id` - The ID of image.
* `name` - The name of image.
* `size` - The size of image.
* `type` - The type of image.
* `os_name` - The name of OS.
* `os_type` - The type of OS.
* `status` - The status of image. Possible values are `Available`, `Making` and `Unavailable`.
* `func_type` - The category of the image, possible values are `gpu`, `app` and `uhost`.
* `scene_categories` - The scene categories of the image, such as `Featured`, `PreInstalledDrivers`, `AIPainting`, `AIModels` and `HPC`.
* `integrated_software` - The name of integrated software (only returned by industry images).
* `vendor` - The vendor of image (only returned by industry images).
* `price_set` - It is a nested type which documented below, the price of industry images. It is empty unless `include_price` is `true`.

The attribute (`price_set`) supports the following:

* `charge_type` - The charge type of image price.
* `original_price` - The original price before discount, in CNY (Chinese Yuan).
* `price` - The price of image, in CNY (Chinese Yuan), rounded to two decimal places.
