---
subcategory: "UMongoDB"
layout: "ucloud"
page_title: "UCloud: ucloud_umongodb_replicaset"
description: |-
  Provides a UMongoDB replica set resource.
---

# ucloud_umongodb_replicaset

Provides a UMongoDB replica set. Node count changes replace the cluster. Disk space and machine type resize the cluster in place.

## Example Usage

```hcl
resource "ucloud_umongodb_replicaset" "example" {
  availability_zone = "cn-bj2-05"
  name              = "tf-umongodb-rs"
  db_version        = "MongoDB 4.2"
  admin_password    = var.admin_password
  disk_space        = 20
  machine_type_id   = "o.mongo2m.medium"
  node_count        = 3
  vpc_id            = ucloud_vpc.default.id
  subnet_id         = ucloud_subnet.default.id
}
```

## Argument Reference

* `availability_zone` - (Required, ForceNew) Availability zone.
* `name` - (Required) Cluster name, 6 to 63 characters.
* `db_version` - (Required, ForceNew) MongoDB version, for example `MongoDB 4.2`.
* `admin_password` - (Required, Sensitive) Administrator password. Updates are sent as raw-url base64. The cluster is started first when it is stopped, and stopped again when `status` stays `stopped`.
* `disk_space` - (Required) Data disk size in GB. Multiple of 10, from 20 to 32000. Changes call `ResizeUMongoDBInstance`.
* `machine_type_id` - (Required) Data node machine type. Changes call `ResizeUMongoDBInstance`.
* `node_count` - (Required, ForceNew) Replica set node count. Only 3, 5, or 7.
* `listen_port` - (Optional, ForceNew) MongoDB listen port.
* `vpc_id` - (Optional, Computed, ForceNew) VPC ID. Read from the cluster when the API returns it.
* `subnet_id` - (Optional, Computed, ForceNew) Subnet ID. Read from the cluster when the API returns it.
* `tag` - (Optional, Computed, ForceNew) Business group. Read from the cluster when the API returns it.
* `charge_type` - (Optional, ForceNew) `year`, `month`, `dynamic`, or `trial`. Default is `month`. The read API does not return it. An empty imported value matches `month` and does not replace the cluster.
* `quantity` - (Optional, ForceNew) Purchase duration. Default is 1. The read API does not return it. An empty imported value matches `1` and does not replace the cluster.
* `cross_zones` - (Optional, Computed, ForceNew) Extra availability zones for a cross-zone replica set. This is a set, so order does not matter. Not available on a sharded cluster.
* `label` - (Optional, ForceNew) Resource labels. Each block has `key` and `value`.
* `sec_group` - (Optional, ForceNew) Security groups. Each block has `id` and `priority` from 1 to 5.
* `status` - (Optional) `running` or `stopped`. Default is `running`.
* `restart_trigger` - (Optional) Changing this value from one non-empty value to another restarts the cluster when `status` is `running`. The read API does not return it, and the first value after import does not restart the cluster.
* `auto_backup_week` - (Optional) Automatic backup weekdays, for example `3,7`. Set this together with `auto_backup_time` unless `backup_disabled` is true.
* `auto_backup_time` - (Optional) Automatic backup window, for example `05:00~06:00`. Set this together with `auto_backup_week` unless `backup_disabled` is true.
* `backup_disabled` - (Optional) Disable automatic backup when true. The weekday and window are not stored while backup is disabled.

## Attributes Reference

* `id` - Cluster ID.
* `state` - Remote cluster state.
* `connect_url` - Connection address.
* `create_time` - Creation time as a Unix timestamp.

## Import

```
$ terraform import ucloud_umongodb_replicaset.example umongodb-rs-abc
```

`admin_password`, `listen_port`, `label`, `sec_group`, and `restart_trigger` are not returned by the read API. Leave `admin_password`, `listen_port`, `label`, and `sec_group` unset after import, or the next apply replaces the cluster. `charge_type` and `quantity` can stay unset: empty imported values match the defaults `month` and `1`. A different charge or quantity must be written in the configuration. Destroy stops a running cluster before deleting it. A cluster left in `InitFailed` or `StartFailed` after its replicas are deleted is removed from state.
