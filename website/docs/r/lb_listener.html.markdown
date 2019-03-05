---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_listener"
sidebar_current: "docs-ucloud-resource-lb-listener"
description: |-
  Provides a Load Balancer Listener resource.
---

# ucloud_lb_listener

Provides a Load Balancer Listener resource.

## Example Usage

```hcl
resource "ucloud_lb" "web" {
    name = "tf-example-lb"
    tag  = "tf-example"
}

resource "ucloud_lb_listener" "example" {
    load_balancer_id = "${ucloud_lb.web.id}"
    protocol         = "https"
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of load balancer instance.
* `protocol` - (Required) Listener protocol. Possible values: `http`, `https` if `listen_type` is `request_proxy`, `tcp` and `udp` if `listen_type` is `packets_transmit`.
* `name` - (Optional) The name of the listener. If not specified, terraform will autogenerate a name beginning with `tf-lb-listener`.
* `listen_type` - (Optional) The type of listener. Possible values are `request_proxy` and `packets_transmit`. (Default: `packets_transmit`).
* `port` - (Optional) Port opened on the listeners to receive requests, range: 1-65535. (Default: `80`).
* `idle_timeout` - (Optional) Amount of time in seconds to wait for the response for in between two sessions if `listen_type` is `request_proxy`, range: 0-86400. (Default: `60`). Amount of time in seconds to wait for one session if `listen_type` is `packets_transmit`, range: 60-900. The session will be closed as soon as no response if it is `0`.
* `method` - (Optional) The load balancer method in which the listener is. Possible values are: `roundrobin`, `source`, `consistent_hash`, `source_port` , `consistent_hash_port`, `weight_roundrobin` and `leastconn`. (Default: `roundrobin`).
    - The `consistent_hash`, `source_port` , `consistent_hash_port`, `roundrobin`, `source` and `weight_roundrobin` are valid if `listen_type` is `packets_transmit`.
    - The `rundrobin`, `source` and `weight_roundrobin` and `leastconn` are vaild if `listen_type` is `request_proxy`.
* `persistence` - (Optional) Indicate whether the persistence session is enabled, it is invaild if `PersistenceType` is `none`, an auto-generated string will be exported if `persistence_type` is `server_insert`, a custom string will be exported if `persistence_type` is `user_defined`.
* `persistence_type` - (Optional) The type of session persistence of listener. Possible values are: `none` as disabled, `server_insert` as auto-generated string and `user_defined` as cutom string. (Default: `none`).
* `health_check_type` - (Optional) Health check method. Possible values are `port` as port checking and `path` as http checking.
* `path` - (Optional) Health check path checking.
* `domain` - (Optional) Health check domain checking.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `status` - Listener status. Possible values are: `allNormal` for all resource functioning well, `partNormal` for partial resource functioning well and `allException` for all resource functioning exceptional.
`