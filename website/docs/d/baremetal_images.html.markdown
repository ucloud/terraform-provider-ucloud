---
subcategory: "Bare Metail (UPhost)"
layout: "ucloud"
page_title: "UCloud: ucloud_baremetal_images"
sidebar_current: "docs-ucloud-datasource-baremetal-images"
description: |-
  Provides a list of bare metal images available in UCloud.

---
# ucloud_baremetal_images

The `ucloud_baremetal_images` data source provides a list of bare metal images available in UCloud. The images can be filtered by their properties.

## Example Usage

```hcl
data "ucloud_baremetal_images" "example" {
  availability_zone = "cn-bj2-02"
  image_type = "base"
  os_type = "centos"
  instance_type = "n-metal-2"
}

output "image_id" {
  value = data.ucloud_baremetal_images.example.images0].id
}

```

## Argument Reference

The following arguments are supported:

- `availability_zone` - (Optional) The availability zone where the images are located.
- `image_type` - (Optional) The type of image. Possible values are `base` for standard images and `custom` for custom images.
- `os_type` - (Optional) The type of OS. Possible values are `centos`, `ubuntu`, and `windows`.
- `instance_type` - (Optional) Machine type of the metal instance where the images can be installed.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `images` - A list of images that match the given criteria. Each image has the following attributes:
  - `availability_zone` - The availability zone where the image is located.
  - `description` - The description of the image, if any.
  - `id` - The ID of the image.
  - `name` - The name of the image.
  - `type` - The type of the image.
  - `os_name` - The name of the OS.
  - `os_type` - The type of the OS.
- `total_count` - The total number of images that satisfy the given criteria.