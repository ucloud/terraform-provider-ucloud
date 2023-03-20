---
subcategory: "Anti-DDoS"
layout: "ucloud"
page_title: "UCloud: ucloud_anti_ddos_ips"
description: |-
  Provides a list of IP from an Anti-DDoS instance.
---

# ucloud_anti_ddos_ips

Provides a list of IP from an Anti-DDoS instance.

## Example Usage

```hcl
data "ucloud_anti_ddos_ips" "ips" {
  instance_id = "usecure_ghp-xxx"
  output_file = "ips.json"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) ID of an Anti-DDoS instance.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ips` - A list of IPs. Each element contains the following attributes
  * `instance_id` - ID of an Anti-DDoS instance.
  * `domain` - Corresponding domain of the IP.
  * `status` - Status of the IP. Possible values are `Pending` and `Success`
  * `comment` - Comment of the IP.
  * `ip` - IP address
  * `proxy_ips` - List of proxy IPs, which should be allowed in firewall or security group policy. Each element is a string.
* `total_count` - Total number of Anti-DDoS instances that satisfy the condition.
