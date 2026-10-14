---
subcategory: "UMongoDB"
layout: "ucloud"
page_title: "UCloud: ucloud_umongodb_instances"
description: |-
  List UMongoDB clusters.
---

# ucloud_umongodb_instances

Lists UMongoDB replica sets and sharded clusters in the provider project.

## Example Usage

```hcl
data "ucloud_umongodb_instances" "example" {
  availability_zone = "cn-bj2-05"
}
```

## Argument Reference

* `availability_zone` - (Optional) Filter by zone. Applied to the returned list.
* `cluster_id` - (Optional) Filter by cluster ID. Applied to the returned list. The current list API request does not keep `ClusterId`, so the provider filters locally.

## Attributes Reference

* `clusters` - List of clusters. Each item has `id`, `name`, `cluster_type`, `db_version`, `disk_space`, `state`, `connect_url`, `zone`, `subnet_id`, `vpc_id`, `tag`, and `create_time`.
