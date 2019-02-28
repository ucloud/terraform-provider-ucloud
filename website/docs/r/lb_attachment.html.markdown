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
resource "ucloud_lb" "web" {
    name = "tf-example-lb"
    tag  = "tf-example"
}

resource "ucloud_lb_listener" "default" {
    load_balancer_id = "${ucloud_lb.web.id}"
    protocol         = "https"
}

resource "ucloud_security_group" "default" {
    name = "tf-example-eip"
    tag  = "tf-example"

    rules {
        port_range = "80"
        protocol   = "tcp"
        cidr_block = "192.168.0.0/16"
        policy     = "accept"
    }
}

resource "ucloud_instance" "web" {
    instance_type     = "n-standard-1"
    availability_zone = "cn-bj2-02"

    root_password      = "wA1234567"
    image_id           = "uimage-of3pac"
    security_group     = "${ucloud_security_group.default.id}"

    name              = "tf-example-lb"
    tag               = "tf-example"
}

resource "ucloud_lb_attachment" "example" {
    load_balancer_id = "${ucloud_lb.web.id}"
    listener_id      = "${ucloud_lb_listener.default.id}"
    resource_id      = "${ucloud_instance.web.id}"
    port             = 80
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of load balancer instance.
* `listener_id` - (Required) The ID of listener servers.
* `resource_id` - (Required) The ID of backend servers.
* `resource_type` - **Deprecated**, attribute `resource_type` is deprecated for optimizing parameters.
* `port` - (Optional) Port opened on the backend server to receive requests, range: 1-65535, (Default: `80`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `private_ip` - The private ip address for backend servers.
* `status` - The status of backend servers. Possible values are: `normalRunning`, `exceptionRunning`.
