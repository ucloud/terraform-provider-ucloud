---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_listeners"
sidebar_current: "docs-ucloud-datasource-lb-listeners"
description: |-
  Provides a list of Load Balancer Listener resources in the current region.
---

# ucloud_lb_listeners

This data source provides a list of Load Balancer Listener resources according to their Load Balancer Listener ID.

## Example Usage

```hcl
data "ucloud_lb_listeners" "example" {
  load_balancer_id = "ulb-xxx"
}

output "first" {
  value = data.ucloud_lb_listeners.example.lb_listeners[0].id
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of a load balancer.

- - -

* `ids` - (Optional) A list of LB Listener IDs, all the LB Listeners belong to this region will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter resulting lb listeners by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `lb_listeners` - It is a nested type which documented below.
* `total_count` - Total number of LB listeners that satisfy the condition.

- - -

The attribute (`lb_listeners`) support the following:

* `id` - The ID of LB Listener.
* `name` - The name of LB Listener.
* `protocol` - LB Listener protocol. Possible values: `http`, `https` if `listen_type` is `request_proxy`, `tcp` and `udp` if `listen_type` is `packets_transmit`.
* `listen_type` - The type of LB Listener. Possible values are `request_proxy` and `packets_transmit`.
* `port` - Port opened on the LB Listener to receive requests, range: 1-65535.
* `idle_timeout` - Amount of time in seconds to wait for the response for in between two sessions if `listen_type` is `request_proxy`, range: 0-86400. Amount of time in seconds to wait for one session if `listen_type` is `packets_transmit`, range: 60-900. The session will be closed as soon as no response if it is `0`.
* `method` - The load balancer method in which the listener is. Possible values are: `roundrobin`, `source`, `consistent_hash`, `source_port` , `consistent_hash_port`, `weight_roundrobin` and `leastconn`. 
    - The `consistent_hash`, `source_port` , `consistent_hash_port`, `roundrobin`, `source` and `weight_roundrobin` are valid if `listen_type` is `packets_transmit`.
    - The `rundrobin`, `source` and `weight_roundrobin` and `leastconn` are vaild if `listen_type` is `request_proxy`.
* `persistence` - Indicate whether the persistence session is enabled, it is invaild if `persistence_type` is `none`, an auto-generated string will be exported if `persistence_type` is `server_insert`, a custom string will be exported if `persistence_type` is `user_defined`.
* `persistence_type` - The type of session persistence of LB Listener. Possible values are: `none` as disabled, `server_insert` as auto-generated string and `user_defined` as cutom string. (Default: `none`).
* `health_check_type` - Health check method. Possible values are `port` as port checking and `path` as http checking.
* `path` - Health check path checking.
* `domain` - Health check domain checking.
* `status` - LB Listener status. Possible values are: `allNormal` for all resource functioning well, `partNormal` for partial resource functioning well and `allException` for all resource functioning exceptional.