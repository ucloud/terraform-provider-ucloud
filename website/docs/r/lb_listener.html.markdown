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
    protocol         = "HTTPS"
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of load balancer instance.
* `protocol` - (Required) Listener protocol. Possible values: HTTP, HTTPS if if "ListenType" is "RequestProxy", TCP and UDP if "ListenType" is "PacketsTransmit".
* `name` - (Optional) The name of the listener, default is "Listener".
* `listen_type` - (Optional) The type of listener, possible values are "RequestProxy" and "PacketsTransmit", default is "PacketsTransmit".
* `port` - (Optional) Port opened on the listeners to receive requests, range from 1 to 65535, and default is 80.
* `idle_timeout` - (Optional) Amount of time in seconds to wait for the response for in between two sessions if "ListenType" is "RequestProxy", range from 0 to 86400 seconds and default is 60. Amount of time in seconds to wait for one session if "ListenType" is "PacketsTransmit", range from 60 to 900, the session will be closed as soon as no response if it is 0.
* `method` - (Optional) The load balance method in which the listener is, possible values are: "Roundrobin", "Source", "ConsistentHash", "SourcePort" and "ConsistentHashPort" . The "ConsistentHash", "SourcePort" and "ConsistentHashPort" is only valid if "listen_type" is "PacketsTransmit" and "Roundrobin", "Source" is vaild if "listen_type" is "RequestProxy" or "PacketsTransmit". Default is "Roundrobin".
* `persistence` - (Optional) Indicate whether the persistence session is enabled, it is invaild if "PersistenceType" is "None", an auto-generated string will be exported if "PersistenceType" is "ServerInsert", a custom string will be exported if "PersistenceType" is "UserDefined".
* `persistence_type` - (Optional) The type of session persistence of listener, it is disabled by default. Possible values are: "None" as disabled, "ServerInsert" as auto-generated string and "UserDefined" as cutom string.
* `health_check_type` - (Optional) Health check method, possible values are "Port" as port checking and "Path" as http checking.
* `path` - (Optional) Health check path checking
* `domain` - (Optional) Health check domain checking

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `status` - Listener status. Possible values are: "AllNormal" as all resource functioning well, "PartNormal" as partial resource functioning well and "AllException" as all resource functioning exceptional.
