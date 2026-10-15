---
subcategory: "UMongoDB"
layout: "ucloud"
page_title: "UCloud: ucloud_umongodb_instance"
description: |-
  Get a UMongoDB cluster.
---

# ucloud_umongodb_instance

Reads one UMongoDB replica set or sharded cluster.

## Example Usage

```hcl
data "ucloud_umongodb_instance" "example" {
  cluster_id = ucloud_umongodb_replicaset.example.id
}
```

## Argument Reference

* `cluster_id` - (Required) Cluster ID.
* `availability_zone` - (Optional, Computed) Availability zone. Read fills this when it is omitted.
* `cluster_type` - (Optional, Computed) `ReplicaSet` or `SharedCluster`. Read fills this when it is omitted.

## Attributes Reference

* `name` - Cluster name.
* `db_version` - MongoDB version.
* `disk_space` - Data disk size in GB.
* `machine_type_id` - Data node machine type.
* `node_count` - Node count, or node count per shard for a sharded cluster.
* `shard_count` - Shard count. Describe sets this from the data replica list, so a replica set is usually 1.
* `mongos_count` - Mongos count. Zero for a replica set.
* `state` - Cluster state.
* `connect_url` - Connection address.
* `subnet_id` - Subnet ID.
* `vpc_id` - VPC ID.
* `tag` - Business group.
* `create_time` - Creation time as a Unix timestamp.
