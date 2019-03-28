---
layout: "ucloud"
page_title: "UCloud: ucloud_lb_attachments"
sidebar_current: "docs-ucloud-datasource-lb-attachments"
description: |-
  Provides a list of Load Balancer Attachment resources under the Load Balancer listener.
---

# ucloud_lb_attachments

This data source provides a list of Load Balancer Attachment resources according to their Load Balancer Attachment ID.

## Example Usage

```hcl
data "ucloud_lb_attachments" "example" {
    load_balancer_id = "ulb-xxx"
    listener_id = "vserver-xxx"
}

output "first" {
    value = "${data.ucloud_lb_attachments.example.lb_attachments.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of a load balancer.
* `listener_id` - (Required) The ID of a listener server.
* `ids` - (Optional) A list of LB Attachment IDs, all the LB Attachments belong to the Load Balancer listener will be retrieved if the ID is `""`.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `lb_attachments` - It is a nested type which documented below.
* `total_count` - Total number of LB Attachments that satisfy the condition.

The attribute (`lb_attachments`) support the following:

* `id` - The ID of LB Attachment.
* `resource_id` - The ID of a backend server.
* `port` - Port opened on the backend server to receive requests, range: 1-65535.
* `private_ip` - The private ip address for backend servers.
* `status` - The status of backend servers. Possible values are: `normalRunning`, `exceptionRunning`.