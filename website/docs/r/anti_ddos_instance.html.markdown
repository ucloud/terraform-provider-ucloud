---
subcategory: "Anti-DDoS"
layout: "ucloud"
page_title: "UCloud: ucloud_anti_ddos_instance"
description: |-
  Provides an Anti-DDoS instance resource in Mainland China.
---

# ucloud_anti_ddos_instance

Provides an Anti-DDoS instance resource in Mainland China..

## Example Usage

```hcl
resource "ucloud_anti_ddos_instance" "foo" {
    area               = "EastChina"
    bandwidth          = 50
    base_defence_value = 30
    data_center        = "Zaozhuang"
    max_defence_value  = 30
    name               = "tf-acc-anti-ddos-instance-basic"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  The name of ucloud_anti_ddos_instance resource, should have 1-63 characters and only support Chinese, English, numbers, '-', '_'.
* `area` - (Required, ForceNew) The area where the instance is deployed. The value can be `EastChina` or `NorthChina`.
* `data_center` - (Required, ForceNew) The data center where the instance is deployed. The value can be `Zaozhuang`, `Yangzhou` or `Taizhou` for `EastChina` area, and `Shijiazhuang` for `NorthChina` area.
* `bandwidth` - (Required) Size of the service bandwidth, whose unit is Mbps.
* `base_defence_value` - (Required) Size of the base defence bandwidth, whose unit is Gbps and minimum value is 30.
* `max_defence_value` - (Required) Size of the maximum defence bandwidth, whose unit is Gbps and value cannot be less than base_defence_value.
* `charge_type` - (Optional, ForceNew) The charge type of Anti-DDoS instance, possible values are year and month (Default: month).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance (Default: 1).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource ucloud_anti_ddos_instance.
* `create_time` - The creation time of ucloud_anti_ddos_instance, formatted in RFC3339 time string.
* `expire_time` - The expiration time of ucloud_anti_ddos_instance, formatted in RFC3339 time string.
* `status` -  The status of ucloud_anti_ddos_instance. Possible values are `Started`, `Stopped` and `Expired`.

## Import

Anti-DDoS instance can be imported using the `id`, e.g.

```
$ terraform import ucloud_anti_ddos_instance.example usecure_ghp-xxx
```
