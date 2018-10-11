---
layout: "ucloud"
page_title: "UCloud: ucloud_images"
sidebar_current: "docs-ucloud-datasource-images"
description: |-
  Provides a list of available image resources in the current region.
---

# ucloud_images

This data source providers a list available image resources according to their availability zone, image ID and other fields.

## Example Usage

```hcl
data "ucloud_images" "example" {
    availability_zone = "cn-sh2-02"
    image_type = "Base"
    os_type = "Linux"
}

output "first" {
    value = "${data.ucloud_images.example.images.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional)Availability zone where instances are located. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `image_id` - (Optional) The ID of image.
* `image_type` - (Optional) The type of image, possible values are: "Base" as standard image, "Business" as owned by market place , and "Custom" as custom-image, all the image types will be retrieved by default.
* `os_type` - (Optional) The type of OS, possible values are: "Linux" and "Windows", all the OS types will be retrieved by default.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `images` - images is a nested type. images documented below.
* `total_count` - Total number of image that satisfy the condition.

The attribute (`images`) support the following:

* `create_time` - The time of creation for EIP.
* `features` - To identify if any particular feature belongs to the instance, the value is "NetEnhnced" as I/O enhanced instance for now.
* `description` - The description of image if any.
* `id` - The ID of image.
* `name` - The name of image.
* `size` - The size of image.
* `type` - The type of image, possible values are: "Base" as standard image, "Business" as owned by market place , and "Custom" as custom-image, all the image types will be retrieved by default.
* `os_name` - The name of OS.
* `os_type` - The type of OS, possible values are: "Linux" and "Windows", all the OS types will be retrieved by default.
* `status` - The status of image, possible values are "Available", "Making" and "Unavailable".
