---
subcategory: "Cube"
layout: "ucloud"
page_title: "UCloud: ucloud_uhub_repository_image_tags"
description: |-
  Provides a list of UHub Repository Image tags.
---

# ucloud_repo_images

This data source provides a list of UHub Repository Image tags according to repository name and image name.

## Example Usage

```hcl
data "ucloud_uhub_repository_image_tags" "foo" {
  repository_name   = "ucloud"
}

output "first" {
  value = data.ucloud_uhub_repository_image_tags.example.repository_image_tags[0].name
}
```

## Argument Reference

The following arguments are supported:

* `image_name` - (Required) The image name.
* `repository_name` - (Required) Image repository name. Possible values are `ucloud`, `cube_lab`, `library` and customize repo name.
* `name` - (Required) The tag name of image.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `repository_image_tags` - It is a nested type which documented below.
* `total_count` - Total number of image tag that satisfy the condition.

- - -

The attribute (`repository_image_tags`) support the following:

* `name` - The tag name of image.
* `digest` - The digest of image.