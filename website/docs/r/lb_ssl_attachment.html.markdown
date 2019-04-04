---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_ssl_attachment"
sidebar_current: "docs-ucloud-resource-lb-ssl-attachment"
description: |-
  Provides a Load Balancer SSL attachment resource for attaching SSL certificate to Load Balancer Listener.
---

# ucloud_lb_ssl

Provides a Load Balancer SSL attachment resource for attaching SSL certificate to Load Balancer Listener.

## Example Usage

```hcl
resource "ucloud_lb" "foo" {
    name = "tf-example-lb-ssl-attachment"
    tag  = "tf-example"
}

resource "ucloud_lb_listener" "foo" {
    name             = "tf-example-lb-ssl-attachment"
    load_balancer_id = "${ucloud_lb.foo.id}"
    protocol         = "https"
}

resource "ucloud_lb_ssl" "foo" {
    name = "tf-example-lb-ssl-attachment"
    private_key = "${file("test-fixtures/private.key")}"
    user_cert = "${file("test-fixtures/user.crt")}"
    ca_cert = "${file("test-fixtures/ca.crt")}"
}

resource "ucloud_lb_ssl_attachment" "foo" {
    load_balancer_id = "${ucloud_lb.foo.id}"
    listener_id      = "${ucloud_lb_listener.foo.id}"
    ssl_id      = "${ucloud_lb_ssl.foo.id}"
}
```

## Argument Reference

The following arguments are supported:

* `ssl_id` - (Required) The ID of SSL certificate.
* `load_balance_id` - (Required) The ID of load balancer instance.
* `listener_id` - (Required)  The ID of listener servers.