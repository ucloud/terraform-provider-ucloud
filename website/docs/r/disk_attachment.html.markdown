---
layout: "ucloud"
page_title: "UCloud: ucloud_disk_attachment"
sidebar_current: "docs-ucloud-resource-disk-attachment"
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
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

# Create security group
resource "ucloud_security_group" "default" {
  name = "tf-example-disk"
  tag  = "tf-example"

  # allow all access from WAN
  rules {
    port_range = "1-65535"
    protocol   = "tcp"
    cidr_block = "0.0.0.0/0"
    policy     = "accept"
  }
}

# Create security group
resource "ucloud_disk" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name              = "tf-example-disk"
  disk_size         = 10
}

# Create a web server
resource "ucloud_instance" "web" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  instance_type     = "n-standard-1"

  image_id      = "${data.ucloud_images.default.images.0.id}"
  root_password = "${var.instance_password}"

  # this security group allows all access from WAN
  security_group = "${ucloud_security_group.default.id}"

  name = "tf-example-disk"
  tag  = "tf-example"
}

# attach disk to instance
resource "ucloud_disk_attachment" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  disk_id           = "${ucloud_disk.default.id}"
  instance_id       = "${ucloud_instance.web.id}"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) The Zone to attach the disk in.
* `instance_id` - (Required) The ID of host instance.
* `disk_id` - (Required) The ID of disk that needs to be attached