---
subcategory: "UHost"
layout: "ucloud"
page_title: "UCloud: ucloud_disk_attachment"
description: |-
  Provides a Cloud Disk Attachment resource for attaching Cloud Disk to UHost Instance.
---

# ucloud_disk_attachment

Provides a Cloud Disk Attachment resource for attaching Cloud Disk to UHost Instance.

## Example Usage

```hcl
# Query availability zone
data "ucloud_zones" "default" {}

# Query image
data "ucloud_images" "default" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

# Create cloud disk
resource "ucloud_disk" "default" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-example-disk"
  disk_size         = 10
}

# Create a web server
resource "ucloud_instance" "web" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  instance_type     = "n-basic-2"

  image_id      = data.ucloud_images.default.images[0].id
  root_password = "wA1234567"

  name = "tf-example-disk"
  tag  = "tf-example"
}

# attach cloud disk to instance
resource "ucloud_disk_attachment" "default" {
  availability_zone              = data.ucloud_zones.default.zones[0].id
  disk_id                        = ucloud_disk.default.id
  instance_id                    = ucloud_instance.web.id
  stop_instance_before_detaching = true
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) The Zone to attach the disk in.
* `instance_id` - (Required, ForceNew) The ID of instance.
* `disk_id` - (Required, ForceNew) The ID of disk that needs to be attached
* `stop_instance_before_detaching` - (Optional, Boolean) Set this to true to ensure that the target instance is stopped
  before trying to detach the volume.