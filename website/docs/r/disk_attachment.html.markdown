---
layout: "ucloud"
page_title: "UCloud: ucloud_disk_attachment"
sidebar_current: "docs-ucloud-resource-disk-attachment"
description: |-
  Provides a Cloud Disk Attachment resource for attachment Cloud Disk to UHost Instance.
---

# ucloud_disk_attachment

Provides a Cloud Disk Attachment resource for attachment Cloud Disk to UHost Instance.

## Example Usage

```hcl
resource "ucloud_disk" "default" {
    availability_zone = "cn-sh2-02"
    name              = "tf-example-disk"
    disk_size         = 10
}

resource "ucloud_security_group" "default" {
    name = "tf-example-eip"
    tag  = "tf-example"

    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

resource "ucloud_instance" "web" {
    instance_type     = "n-standard-1"
    availability_zone = "cn-sh2-02"

    root_password      = "wA1234567"
    image_id           = "uimage-of3pac"
    security_group     = "${ucloud_security_group.default.id}"

    name              = "tf-example-disk"
    tag               = "tf-example"
}

resource "ucloud_disk_attachment" "example" {
    availability_zone = "cn-sh2-02"
    disk_id = "${ucloud_disk.default.id}"
    instance_id = "${ucloud_instance.web.id}"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) The Zone to attach the disk in.
* `instance_id` - (Required) The ID of host instance.
* `disk_id` - (Required) The ID of disk that needs to be attached