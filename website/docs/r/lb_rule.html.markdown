---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_rule"
sidebar_current: "docs-ucloud-resource-lb-rule"
description: |-
  Provides a Load Balancer Rule resource to add content forwarding policies for Load Balancer backend resource.
---

# ucloud_lb_rule

Provides a Load Balancer Rule resource to add content forwarding policies for Load Balancer backend resource.
 
~> **Note** The Load Balancer Rule can only be define while the `protocol` of lb listener is one of HTTP and HTTPS. In addition, should set one of `domain` and `path` if defined.

## Example Usage

```hcl
data "ucloud_images" "default" {
  availability_zone = "cn-bj2-02"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

resource "ucloud_lb" "web" {
  name = "tf-example-lb"
  tag  = "tf-example"
}

resource "ucloud_lb_listener" "default" {
  load_balancer_id = ucloud_lb.web.id
  protocol         = "http"
}

resource "ucloud_instance" "web" {
  instance_type     = "n-basic-2"
  availability_zone = "cn-bj2-02"

  root_password = "wA1234567"
  image_id      = data.ucloud_images.default.images[0].id

  name = "tf-example-lb"
  tag  = "tf-example"
}

resource "ucloud_lb_attachment" "default" {
  load_balancer_id = ucloud_lb.web.id
  listener_id      = ucloud_lb_listener.default.id
  resource_type    = "instance"
  resource_id      = ucloud_instance.web.id
  port             = 80
}

resource "ucloud_lb_rule" "example" {
  load_balancer_id = ucloud_lb.web.id
  listener_id      = ucloud_lb_listener.default.id
  backend_ids      = ucloud_lb_attachment.default.*.id
  domain           = "www.ucloud.cn"
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required, ForceNew) The ID of a load balancer.
* `listener_id` - (Required, ForceNew) The ID of a listener server.
* `backend_ids` - (Required, ForceNew) The IDs of the backend servers where rule applies, this argument is populated base on the `backend_id` responded from `lb_attachment` create.

- - -

* `path` - (Optional) The path of Content forward matching fields. `path` and `domain` cannot coexist. `path` and `domain` must be filled in one.
* `domain` - (Optional) The domain of content forward matching fields. `path` and `domain` cannot coexist. `path` and `domain` must be filled in one.