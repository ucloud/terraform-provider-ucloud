---
subcategory: "Anti-DDoS"
layout: "ucloud"
page_title: "UCloud: ucloud_anti_ddos_instances"
description: |-
  Provides a list of Anti-DDoS instance resources.
---

# ucloud_anti_ddos_instances

Provides a list of Anti-DDoS instance resources.

## Example Usage

```hcl
data "ucloud_anti_ddos_instances" "instance" {
  output_file = "instances.json"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional) A list of Anti-DDoS instance IDs, all the Anti-DDoS instances will be retrieved if the ID is `[]`.
* `name_regex` - (Optional) A regex string to filter result by name.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instances` -  A list of Anti-DDoS instances. Each element contains the following attributes
  * `id` - The ID of the resource.
  * `name` - The name of the resource.
  * `area` - The area where the instance is deployed. The value can be `EastChina` or `NorthChina`.
  * `data_center` - The data center where the instance is deployed. The value can be `Zaozhuang`, `Yangzhou` or `Taizhou` for `EastChina` area, and `Shijiazhuang` for `NorthChina` area.
  * `bandwidth` - Size of the service bandwidth, whose unit is Mbps.
  * `base_defence_value` - Size of the base defence bandwidth, whose unit is Gbps and minimum value is 30.
  * `max_defence_value` - Size of the maximum defence bandwidth, whose unit is Gbps and value cannot be less than base_defence_value.
  * `charge_type` - The charge type of Anti-DDoS instance, possible values are year and month (Default: month).
  * `create_time` - The creation time of ucloud_anti_ddos_instance, formatted in RFC3339 time string.
  * `expire_time` - The expiration time of ucloud_anti_ddos_instance, formatted in RFC3339 time string.
  * `status` -  The status of ucloud_anti_ddos_instance. Possible values are `Started`, `Stopped` and `Expired`.
* `total_count` - Total number of Anti-DDoS instances that satisfy the condition.


