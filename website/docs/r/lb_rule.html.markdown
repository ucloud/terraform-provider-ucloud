---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_rule"
sidebar_current: "docs-ucloud-resource-lb-rule"
description: |-
  Provides a Load Balancer Rule resource to add content forwarding policies for Load Balancer backend resource.
---

# ucloud_lb_rule

Provides a Load Balancer Rule resource to add content forwarding policies for Load Balancer backend resource.

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

resource "ucloud_lb_attachment" "default" {
    load_balancer_id = "${ucloud_lb.web.id}"
    listener_id      = "${ucloud_lb_listener.default.id}"
    resource_type    = "instance"
    resource_id      = "${ucloud_instance.web.id}"
    port             = 80
}

resource "ucloud_lb_rule" "example" {
    load_balancer_id = "${ucloud_lb.web.id}"
    listener_id      = "${ucloud_lb_listener.default.id}"
    backend_ids      = ["${ucloud_lb_attachment.default.id}"]
    domain           = "www.ucloud.cn"
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of a load balancer.
* `listener_id` - (Required) The ID of a listener server.
* `backend_ids` - (Required) The IDs of the backend servers where rule applies, this argument is populated base on the `backend_id` responed from `lb_attachment` create.
* `path` - (Optional) The path of Content forward matching fields. `path` and `domain` cannot coexist. `path` and `domain` must be filled in one.
* `domain` - (Optional) The domain of content forward matching fields. `path` and `domain` cannot coexist. `path` and `domain` must be filled in one.