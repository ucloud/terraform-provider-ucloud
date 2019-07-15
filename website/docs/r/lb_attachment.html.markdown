---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_attachment"
sidebar_current: "docs-ucloud-resource-lb-attachment"
description: |-
  Provides a Load Balancer Attachment resource for attaching Load Balancer to UHost Instance, etc.
---

# ucloud_lb_attachment

Provides a Load Balancer Attachment resource for attaching Load Balancer to UHost Instance, etc.

## Example Usage

```hcl
# Query image
data "ucloud_images" "default" {
  availability_zone = "cn-bj2-04"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

# Create Load Balancer
resource "ucloud_lb" "web" {
  name = "tf-example-lb"
  tag  = "tf-example"
}

# Create Load Balancer Listener with http protocol
resource "ucloud_lb_listener" "default" {
  load_balancer_id = ucloud_lb.web.id
  protocol         = "http"
}

# Create web server
resource "ucloud_instance" "web" {
  instance_type     = "n-basic-2"
  availability_zone = "cn-bj2-04"

  root_password = "wA1234567"
  image_id      = data.ucloud_images.default.images[0].id

  name = "tf-example-lb"
  tag  = "tf-example"
}

# Attach instances to Load Balancer
resource "ucloud_lb_attachment" "example" {
  load_balancer_id = ucloud_lb.web.id
  listener_id      = ucloud_lb_listener.default.id
  resource_id      = ucloud_instance.web.id
  port             = 80
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of a load balancer.
* `listener_id` - (Required) The ID of a listener server.
* `resource_id` - (Required) The ID of a backend server.
* `resource_type` - **Deprecated**, attribute `resource_type` is deprecated for optimizing parameters.
* `port` - (Optional) Port opened on the backend server to receive requests, range: 1-65535, (Default: `80`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `private_ip` - The private ip address for backend servers.
* `status` - The status of backend servers. Possible values are: `normalRunning`, `exceptionRunning`.
