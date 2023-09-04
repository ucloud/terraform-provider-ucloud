---
subcategory: "Bare Metal (UPhost)"
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
  os_type = "CentOS"
  instance_type = "xyz-type"
}

output "image_id" {
  value = data.ucloud_baremetal_images.example.images[0].id
}

```

## Argument Reference

The following arguments are supported:

- `availability_zone` - (Optional) The availability zone where the images are located.
- `image_type` - (Optional) The type of image. Possible values are `base` for standard images and `custom` for custom images.
- `os_type` - (Optional) The type of OS. Possible values are `CentOS`, `Ubuntu`, and `Windows`.
- `instance_type` - (Optional) Machine type of the bare metal instance where the images can be installed.
* `ids` - (Optional) A list of image IDs, all the images belong to this region will be retrieved if the ID is `[]`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).
* `name_regex` - (Optional) A regex string to filter resulting images by name. (Such as: `^CentOS 7.[1-2] 64` means CentOS 7.1 of 64-bit operating system or CentOS 7.2 of 64-bit operating system, "^Ubuntu 16.04 64" means Ubuntu 16.04 of 64-bit operating system).

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
  - `status` - The status of image. Possible values are `Available`, `Making` and `Unavailable`.
- `total_count` - The total number of images that satisfy the given criteria.