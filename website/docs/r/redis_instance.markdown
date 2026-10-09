---
subcategory: "UMem"
layout: "ucloud"
page_title: "UCloud: ucloud_redis_instance"
description: |-
  Provides a Redis instance resource.
---

# ucloud_redis_instance

The UCloud Redis instance is a key-value online storage service compatible with the Redis protocol.

## Example Usage

```hcl
data "ucloud_zones" "default" {}

resource "ucloud_vpc" "example" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "example" {
  name       = "tf-example-subnet"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.example.id
}

resource "ucloud_redis_instance" "master" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  instance_type     = "redis-master-2"
  password          = "2018_Tfacc"
  engine_version    = "4.0"
  vpc_id            = ucloud_vpc.example.id
  subnet_id         = ucloud_subnet.example.id

  name = "tf-example-redis-master"
  tag  = "tf-example"
}

resource "ucloud_redis_instance" "distributed" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  instance_type     = "redis-distributed-16"
  block_cnt         = 4
  vpc_id            = ucloud_vpc.example.id
  subnet_id         = ucloud_subnet.example.id

  name = "tf-example-redis-distributed"
  tag  = "tf-example"
}
```

### Restart an Active-Standby Redis instance

Bump `restart_trigger` to restart the instance in place. The value is arbitrary; setting it to a fresh `timestamp()` restarts on every apply where it changes.

```hcl
resource "ucloud_redis_instance" "master" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  instance_type     = "redis-master-2"
  password          = "2018_Tfacc"
  engine_version    = "4.0"
  vpc_id            = ucloud_vpc.example.id
  subnet_id         = ucloud_subnet.example.id

  name = "tf-example-redis-master"
  tag  = "tf-example"

  restart_trigger = "2026-09-07T10:00:00Z"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where Redis instance is located. Such as: "cn-bj2-02". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `instance_type` - (Required) The type of Redis instance, please visit the [instance type table](https://docs.ucloud.cn/terraform/specification/umem_instance?id=redis) for more details.
* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the Redis instance.
* `subnet_id` - (Required, ForceNew) The ID of subnet linked to the Redis instance.

- - -

* `name` - (Optional) The name of Redis instance, which contains 6-63 characters and only support English, numbers, '-', '_'. If not specified, terraform will auto-generate a name beginning with `tf-redis-instance`.
* `standby_zone` - (Optional, ForceNew) Availability zone where the standby Redis instance is located for the high availability Redis instance with multiple zone; only be supported for Active-Standby Redis, not be supported for Distributed Redis.
* `block_cnt` - (Optional, ForceNew) The number of shards (blocks) of a Distributed Redis instance, sent as the `BlockCnt` parameter of `CreateUMemSpace`. Only supported for Distributed Redis; rejected for Active-Standby Redis. If omitted, a default of `4` is used. Changing it forces a new instance, since the shard count cannot be modified in place (a capacity resize keeps the existing shards). The live shards are exposed via the `block_set` attribute.
* `charge_type` - (Optional, ForceNew) The charge type of Redis instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the Redis instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `tag` - (Optional, ForceNew) A tag assigned to Redis instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or an empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `engine_version` - (active-standby Redis Required, ForceNew) The version of engine of active-standby Redis.
* `password` - (Optional) The password for  active-standby Redis instance which should have 6-36 characters. It must contain at least 3 items of Capital letters, small letter, numbers and special characters. The special characters include `-_`.
* `auto_backup` - (Optional) Enable or not start auto backup of Redis instance, only support for Active-Standby Redis, possible values are: `enable`, `disable`.
* `backup_begin_time` - (Optional) Specifies when the backup starts, measured in hour, only support for Active-Standby Redis, it starts at 3 o'clock in the morning by default, possible values are: 0-23.
* `restart_trigger` - (Optional) A trigger used to request a restart of an Active-Standby Redis instance. Changing its value (for example, to a new timestamp) issues a `RestartURedisGroup` call and waits until the instance returns to `Running`. Only supported for Active-Standby Redis; not supported for Distributed Redis.

~> **Note** The active-standby Redis doesn't support to be created on multiple zones with Terraform.

~> **Note** Resizing a Distributed Redis instance grows its shards one by one via the `ResizeUDRedisBlockSize` API. The new total capacity must equal `shard count x per-shard size`, and the per-shard size is restricted to 4/8/12/16/20 GB (for example, a 4-shard instance supports totals of 16/32/48/64/80 GB). Listing the shards requires access to the `DescribeUMemBlockInfo` API.

~> **Note** `restart_trigger` is an imperative, one-shot action: it does nothing on create, and only fires when its value changes between applies. A common pattern is to set it to `timestamp()` or a version string you bump on demand.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource Redis instance.
* `ip_set` - ip_set is a nested type. ip_set documented below.
* `create_time` - The creation time of Redis instance, formatted by RFC3339 time string.
* `expire_time` - The expiration time of Redis instance, formatted by RFC3339 time string.
* `status` - The status of KV Redis instance.

- - -

The attribute (`ip_set`) support the following:

* `ip` - The virtual ip of Redis instance.
* `port` - The port on which Redis instance accepts connections, it is 6379 by default.
