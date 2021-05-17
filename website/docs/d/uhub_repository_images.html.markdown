---
subcategory: "Cube"
layout: "ucloud"
page_title: "UCloud: ucloud_repository_images"
description: |-
  Provides a list of Repository Images.
---

# ucloud_repo_images

This data source provides a list of Repository Images according to repository name.

## Example Usage

```hcl
data "ucloud_repository_images" "foo" {
  repository_name   = "ucloud"
}

output "first" {
  value = data.ucloud_repository_images.example.repository_images[0].name
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional) A regex string to filter resulting Images by image name.
* `repository_name` - (Required) Image repository name. Possible values are `ucloud`, `cube_lab`, `library` and customize repository name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `repository_images` - It is a nested type which documented below.
* `total_count` - Total number of image that satisfy the condition.

- - -

The attribute (`repository_images`) support the following:

* `name` - The name of image.
* `latest_tag` - The latest tag assigned to image.