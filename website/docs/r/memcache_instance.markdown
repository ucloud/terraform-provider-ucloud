---
layout: "ucloud"
page_title: "UCloud: ucloud_memcache_instance"
sidebar_current: "docs-ucloud-resource-memcache-instance"
description: |-
  Provides a Memcache instance resource.
---

# ucloud_memcache_instance

The UCloud Memcache instance is a key-value online storage service compatible with the Memcached protocol.

## Example Usage

```hcl
data "ucloud_zones" "default" {}

resource "ucloud_memcache_instance" "master" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  instance_type     = "memcache-master-2"

  name = "tf-example-memcache"
  tag  = "tf-example"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) Availability zone where Memcache instance is located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `instance_type` - (Required) The type of Memcache instance, please visit the [instance type table](../appendix/memcache_instance_type.html) for more details.

- - -

* `name` - (Optional) The name of Memcache instance, which contains 6-63 characters and only support English, numbers, '-', '_'. If not specified, terraform will auto-generate a name beginning with `tf-memcache-instance`.
* `charge_type` - (Optional) The charge type of Memcache instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional) The duration that you will buy the Memcache instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `tag` - (Optional) A tag assigned to Memcache instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `vpc_id` - (Optional) The ID of VPC linked to the Memcache instance.
* `subnet_id` - (Optional) The ID of subnet linked to the Memcache instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ip_set` - ip_set is a nested type. ip_set documented below.
* `create_time` - The creation time of Memcache instance, formatted by RFC3339 time string.
* `expire_time` - The expiration time of Memcache instance, formatted by RFC3339 time string.
* `status` - The status of KV Memcache instance.

- - -

The attribute (`ip_set`) support the following:

* `ip` - The virtual ip of Memcache instance.
* `port` - The port on which Memcache instance accepts connections, it is 6379 by default.
