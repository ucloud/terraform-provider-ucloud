---
subcategory: "ULB"
layout: "ucloud"
page_title: "UCloud: ucloud_lb_listener"
description: |-
  Provides a Load Balancer Listener resource.
---

# ucloud_lb_listener

Provides a Load Balancer Listener resource.

~> **Note** This `listen_type` only support when `protocol` is `tcp` in the extranet mode and the default value is `request_proxy`. In addition, in the extranet mode, the `listen_type` is `request_proxy` if `protocol`is `http` or `https`, the `listen_type` is `packets_transmit` if `protocol`is `udp`. In the intranet mode, the `listen_type` is `packets_transmit`.

## Example Usage

```hcl
resource "ucloud_lb" "web" {
  name = "tf-example-lb"
  tag  = "tf-example"
}

resource "ucloud_lb_listener" "example" {
  load_balancer_id = ucloud_lb.web.id
  protocol         = "http"
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required, ForceNew) The ID of load balancer instance.
* `protocol` - (Required, ForceNew) Listener protocol. Possible values: `http`, `https`, `tcp` if `listen_type` is `request_proxy`, `tcp` and `udp` if `listen_type` is `packets_transmit`.

- - -

* `name` - (Optional) The name of the listener. If not specified, terraform will auto-generate a name beginning with `tf-lb-listener`.
* `listen_type` - (Optional, ForceNew) The type of listener. Possible values are `request_proxy` and `packets_transmit`. When `packets_transmit` was specified, you need to config the instances by yourself if the instances attach to the load balancer. You may refer to [configuration instruction](https://docs.ucloud.cn/network/ulb/fast/createulb/vservertype).
* `port` - (Optional, ForceNew) Port opened on the listeners to receive requests, range: 1-65535. The default value: `80` as `protocol` is `http`, `443` as `protocol` is `https`, `1024` as `protocol` is `tcp` or `udp`.
* `idle_timeout` - (Optional) Keep alive timeout of the connection between the client and LB, measured in second. Range: 0-86400 when `listen_type` is `request_proxy`, range: 60-900 when `listen_type` is `packets_transmit` (Default: `60`). The connection will be closed as soon as no response between the client and LB if it set by `0`.
* `method` - (Optional) The load balancer method in which the listener is. Possible values are: `roundrobin`, `source`, `consistent_hash`, `source_port` , `consistent_hash_port`, `weight_roundrobin` and `leastconn`. (Default: `roundrobin`).
    - The `consistent_hash`, `source_port` , `consistent_hash_port`, `roundrobin`, `source` and `weight_roundrobin` are valid if `listen_type` is `packets_transmit`.
    - The `roundrobin`, `source` and `weight_roundrobin` and `leastconn` are valid if `listen_type` is `request_proxy`.
* `persistence` - (Optional) Indicate whether the persistence session is enabled, it is invalid if `persistence_type` is `none`, an auto-generated string will be exported if `persistence_type` is `server_insert`, a custom string will be exported if `persistence_type` is `user_defined`.
* `persistence_type` - (Optional) The type of session persistence of listener. Possible values are: `none` as disabled, `server_insert` as auto-generated key and `user_defined` as customized key. (Default: `none`).
* `health_check_type` - (Optional) Health check method. Possible values are `port` as port checking and `path` as http checking.
* `path` - (Optional) Health check path checking.
* `domain` - (Optional) Health check domain checking.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource lb listener.
* `status` - Listener status. Possible values are: `allNormal` for all resource functioning well, `partNormal` for partial resource functioning well and `allException` for all resource functioning exceptional.

## Import

LB Listener can be imported using the `id`, e.g.

```
$ terraform import ucloud_lb_listener.example vserver-abcdefg
```