---
subcategory: "UK8S"
layout: "ucloud"
page_title: "UCloud: ucloud_uk8s_images"
description: |-
  Provides a list of UK8S available images in the current region.
---

# ucloud_uk8s_images

This data source provides a list of images supported by UK8S, which can be used to create UK8S clusters and nodes. Regular `ucloud_images` cannot be used for UK8S, use this data source instead.

## Example Usage

```hcl
data "ucloud_uk8s_images" "default" {
  availability_zone = "cn-bj2-02"
  name_regex        = "^Ubuntu"
}

output "first_image_id" {
  value = data.ucloud_uk8s_images.default.images.0.id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability zone where images are located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `host_type` - (Optional) The host type of the images. Possible values are: `uhost` for cloud host images, `uphost` for physical cloud host images. (Default: `uhost`).
* `name_regex` - (Optional) A regex string to filter resulting images by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `total_count` - Total number of images that satisfy the condition.
* `images` - It is a nested type which documented below.

The attribute (`images`) supports the following:

* `id` - The ID of image.
* `name` - The name of image.
* `zone_id` - The ID of the availability zone where the image is located.
* `not_support_gpu` - Whether the image does not support GPU machine type.
