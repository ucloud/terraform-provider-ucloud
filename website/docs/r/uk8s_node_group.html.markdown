---
subcategory: "UK8S"
layout: "ucloud"
page_title: "UCloud: ucloud_uk8s_node_group"
sidebar_current: "docs-ucloud-resource-uk8s-node-group"
description: |-
  Provides a UK8S node group resource.
---

# ucloud_uk8s_node_group

Provides a UK8S node group. Labels and taints are stored in the node group template and are used by nodes subsequently added to this group. Updating them changes the node group in place and does not recreate existing nodes.

## Example Usage

```hcl
resource "ucloud_uk8s_node_group" "workers" {
  cluster_id        = ucloud_uk8s_cluster.foo.id
  name              = "workers"
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "cn-sh2-02"
  image_id           = "uimage-1rgr4qmrvmjm"
  instance_type      = "o-basic-2"
  uhost_family       = "o2i"
  charge_type        = "dynamic"
  boot_disk_type     = "cloud_rssd"
  data_disk_type     = "cloud_rssd"
  data_disk_size     = 100

  labels = {
    environment = "test"
    role        = "worker1"
  }

  taint {
    key    = "dedicated"
    value  = "test"
    effect = "NoSchedule"
  }
}

resource "ucloud_uk8s_node" "workers" {
  count = 2

  cluster_id        = ucloud_uk8s_cluster.foo.id
  node_group_id     = ucloud_uk8s_node_group.workers.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "cn-sh2-02"
  password          = var.password
  image_id           = "uimage-1rgr4qmrvmjm"
  instance_type      = "o-basic-2"
  charge_type        = "dynamic"
  boot_disk_type     = "cloud_rssd"
}
```

## Argument Reference

* `cluster_id` - (Required, ForceNew) The ID of the UK8S cluster.
* `name` - (Required) The name of the node group.
* `subnet_id` - (Required) The subnet ID used by the node group template.
* `availability_zone` - (Required) The availability zone used by the node group template.
* `instance_type` - (Required) The worker instance type, such as `o-basic-2`.
* `uhost_family` - (Optional, Computed) The Intel outstanding instance family: `o1i` or `o2i`. Use `instance_type = "o-basic-2"`, `uhost_family = "o2i"`, and `boot_disk_type = "cloud_rssd"` for O2 Intel nodes. For an RSSD data disk, also set `data_disk_type = "cloud_rssd"` and a positive `data_disk_size`. Requires an Intel CPU platform. When omitted, the family is read from UK8S. Updates change the template for future nodes without replacing existing nodes.
* `image_id` - (Optional) The image ID used by new nodes.
* `charge_type` - (Optional) The charge type. Valid values are `year`, `month`, and `dynamic`. Default is `month`.
* `boot_disk_type` - (Optional) The boot disk type, including `cloud_rssd`. Default is `cloud_ssd`. Set `cloud_rssd` when using `uhost_family = "o1i"` or `"o2i"`.
* `boot_disk_size` - (Optional) The boot disk size in GB. Default is `40`.
* `data_disk_type` - (Optional) The data disk type, including `cloud_rssd`. Default is `cloud_ssd`. Set a positive `data_disk_size` to create a data disk.
* `data_disk_size` - (Optional) The data disk size in GB. Default is `0`.
* `isolation_group` - (Optional) The isolation group ID.
* `min_cpu_platform` - (Optional) The minimum CPU platform. Default is `Intel/Auto`.
* `max_pods` - (Optional) The maximum number of pods per node. Default is `110`.
* `tag` - (Optional) The business tag.
* `user_data` - (Optional) User data run while the instance initializes.
* `init_script` - (Optional) A script run after the node initializes.
* `labels` - (Optional) A map of Kubernetes labels for nodes added from this node group. At most 20 labels are supported.
* `taint` - (Optional) One or more taints for nodes added from this node group. At most 10 taints are supported. Each block accepts `key`, `value`, and an `effect` of `NoSchedule`, `PreferNoSchedule`, or `NoExecute`.

## Attributes Reference

* `node_ids` - IDs of nodes currently associated with the node group.
* `create_time` - The node group creation time in RFC3339 format.
* `update_time` - The last node group update time in RFC3339 format.

## Update behavior

Changing `uhost_family`, `instance_type`, `boot_disk_type`, `data_disk_type`, `data_disk_size`, `labels`, or `taint` calls `UpdateUK8SNodeGroup` and keeps the node group ID. The updated template is applied when new nodes are added to the group. Existing nodes are not recreated or modified automatically.

For example, change the group's `uhost_family` from `o1i` to `o2i` and set `data_disk_type = "cloud_rssd"` with a positive `data_disk_size`. Terraform plans an in-place node group update. New nodes created with this `node_group_id` inherit the group's family when the node's `uhost_family` is omitted. UK8S also uses the group's boot and data disk configuration for those nodes.

For nodes created without a node group, set `uhost_family = "o2i"`, `boot_disk_type = "cloud_rssd"`, and, if needed, `data_disk_type = "cloud_rssd"` and `data_disk_size` on `ucloud_uk8s_node` itself. Changing `uhost_family` or disk types on an existing node resource requires replacing that node; updating the node group template does not.

The node group must be empty before it can be deleted.

## Import

Import a node group using its cluster ID and node group ID, separated by `/`:

```sh
terraform import ucloud_uk8s_node_group.workers uk8s-example/nodegroup-example
```

Configure the provider with the cluster's region and project. Import stores `cluster_id` separately and keeps the node group ID as the Terraform resource ID, matching normal creation. The remaining template fields are populated by reading the node group. Add the corresponding resource configuration and review `terraform plan` after import.
