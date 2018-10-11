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
    protocol         = "HTTPS"
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

* `load_balancer_id` - (Required) The ID of the load balancer which requires the rule.
* `listener_id` - (Required) The ID of the listeners which require the rule.
* `backend_ids` - (Required) The ID of the backend server where rule applies , this argument is populated base on the "BackendId" responed from "lb attachment create".
* `path` - (Optional) The path of Content forward matching fields. path and domain cannot coexist. path and domain must fill in one.
* `domain` - (Optional) The domain of Content forward matching fields.path and domain cannot coexist. path and domain must fill in one.