---
subcategory: "UMongoDB"
layout: "ucloud"
page_title: "UCloud: ucloud_umongodb_sharded_cluster"
description: |-
  Provides a UMongoDB sharded cluster resource.
---

# ucloud_umongodb_sharded_cluster

Provides a UMongoDB sharded cluster. Shard count and node count replace the cluster. Disk space and machine type resize it in place.

## Example Usage

```hcl
resource "ucloud_umongodb_sharded_cluster" "example" {
  availability_zone = "cn-bj2-05"
  name              = "tf-umongodb-sc"
  db_version        = "MongoDB 4.2"
  admin_password    = var.admin_password
  disk_space        = 20
  machine_type_id   = "o.mongo2m.medium"
  node_count        = 3
  shard_count       = 2
  mongos_node_count = 2
  vpc_id            = ucloud_vpc.default.id
  subnet_id         = ucloud_subnet.default.id
}
```

## Argument Reference

Arguments match `ucloud_umongodb_replicaset`, except `cross_zones` is not supported, plus:

* `node_count` - (Required, ForceNew) Node count in each shard. Only 3, 5, or 7.
* `shard_count` - (Required, ForceNew) Number of shards. From 2 to 9.
* `mongos_node_count` - (Required, ForceNew) Number of mongos nodes. From 2 to 5.
* `mongos_machine_type_id` - (Optional) Mongos machine type. Changes call `ResizeUMongoDBInstance`.

## Import

```
$ terraform import ucloud_umongodb_sharded_cluster.example umongodb-sc-abc
```
