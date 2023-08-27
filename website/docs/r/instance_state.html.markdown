---
subcategory: "UHost"
layout: "ucloud"
page_title: "UCloud: ucloud_instance_state"
description: |-
  Provides an UHost Instance State resource.
---

# Resource: ucloud_instance_state

Provides an UHost Instance State resource. This allows managing an instance power state.

## Example Usage

```terraform
variable "availability_zone" {
  type    = string
  default = "cn-bj2-05"
}
data "ucloud_images" "default" {
  availability_zone = "${var.availability_zone}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  availability_zone = "${var.availability_zone}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  charge_type       = "month"
  duration          = 0
  name              = "tf-acc-instance-config-basic"
  tag               = "tf-acc"
}
resource "ucloud_instance_state" "foo" {
	instance_id = "${ucloud_instance.foo.id}"
	force  = true
	state = "Stopped"
}
```

## Argument Reference

The following arguments are required:

* `instance_id` - (Required) ID of the instance.
* `state` - (Required) - State of the instance. Valid values are `Stopped`, `Running`.

The following arguments are optional:

* `force` - (Optional) Whether to request a forced stop when `state` is `Stopped`. Otherwise (_i.e._, `State` is `Running`), ignored. When an instance is forced to stop, it does not flush system caches and buffer. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the instance (matches `instance_id`).

## Import

Using `terraform import`, import `ucloud_instance_state` using the `instance_id` attribute. For example:

```console
$ terraform import ucloud_instance_state.test uhost-xyz
```
