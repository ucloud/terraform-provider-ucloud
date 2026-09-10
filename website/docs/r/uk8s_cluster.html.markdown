---
subcategory: "UK8S"
layout: "ucloud"
page_title: "UCloud: ucloud_uk8s_cluster"
description: |-
  Provides an UK8S Cluster resource.
---

# ucloud_uk8s_cluster

Provides an UK8S Cluster resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-uk8s-cluster"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
  name       = "tf-acc-uk8s-cluster"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
  vpc_id       = "${ucloud_vpc.foo.id}"
  subnet_id    = "${ucloud_subnet.foo.id}"
  name         = "tf-acc-uk8s-cluster-basic-update"
  service_cidr = "172.16.0.0/16"
  cni_mode     = "VPC"
  password     = "ucloud_2021"
  charge_type  = "dynamic"

  master {
    availability_zones = [
      "${data.ucloud_zones.default.zones.0.id}",
      "${data.ucloud_zones.default.zones.0.id}",
      "${data.ucloud_zones.default.zones.0.id}",
    ]
    machine_type = "N"
    cpu          = 2
    memory       = 4096
  }
}
```

## Argument Reference

The following arguments are supported:

* `service_cidr` - (Required, ForceNew) The CIDR block of k8s service.
* `cni_mode` - (Optional, Computed, ForceNew) The cluster CNI network mode, `VPC` or `Calico`, sent as `CNIMode`. When omitted, the API selects the mode and Terraform reads it from the cluster. Changing the mode replaces the cluster.
* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the instance. If not defined `vpc_id`, the instance will use the default VPC in the current region.
* `subnet_id` - (Required, ForceNew) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. If not defined `vpc_id` and `subnet_id`, the instance will use the default subnet in the current region.
* `password` - (Required) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will auto-generate a password.

---

* `name` - (Optional) The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will auto-generate a name beginning with `tf-instance`.
* `user_data` - (Optional, ForceNew) The user data to customize the startup behaviors when launching the instance. You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
* `init_script` - (Optional, ForceNew) The user data to customize the startup behaviors when launching the instance. You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
* `charge_type` - (Optional, ForceNew) The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `k8s_version` - (Optional, ForceNew) The version of k8s. See also [Create UK8S](https://console.ucloud.cn/uk8s/create).
* `enable_external_api_server` - (Optional, ForceNew) If expose the api server endpoint for external visiting.
* `delete_disks_with_instance` - (Optional, ForceNew) Whether the cloud data disks attached instance should be destroyed on instance termination.
* `kube_proxy` - (Optional, ForceNew) The configuration of kube proxy, See [kube proxy](#kube_proxy) for details of attributes.
* `master` - (Required) The configuration of master, See [master](#master) for details of attributes.
* `worker` - (Optional, ForceNew) Repeatable initial Worker group configuration, sent as `Nodes.N.*` in `CreateUK8SClusterV2`. See [worker](#worker).
* `image_id` - (Optional, ForceNew) The default image ID. `master.image_id` takes precedence for Master nodes.

### master

The `master` supports the following:

* `availability_zones` - (Required, ForceNew) Availability zone list where instance is located. such as: `["cn-bj2-02", "cn-bj2-03", "cn-bj2-05"]`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `instance_type` - (Optional, Deprecated, ForceNew) The legacy specification, such as `n-basic-2`. Existing configurations remain supported with the original validation rules. Conflicts with `machine_type`, `cpu`, and `memory`.
* `machine_type` - (Optional, ForceNew) The uppercase machine type, such as `N` or `O`. Required together with `cpu` and `memory` when `instance_type` is omitted.
* `cpu` - (Optional, ForceNew) Number of CPU cores, at least 2. Required with `machine_type` and `memory` when `instance_type` is omitted. Available specifications depend on the machine type and region.
* `memory` - (Optional, ForceNew) Memory in MB, at least 4096 and a multiple of 1024. Required with `machine_type` and `cpu` when `instance_type` is omitted. For 4 GB, use `4096`.
* `uhost_family` - (Optional, ForceNew) The Intel outstanding family, `o1i` or `o2i`, sent as `MasterUHostFamily`. Requires `machine_type = "O"`, an Intel CPU platform and `boot_disk_type = "cloud_rssd"`.
* `image_id` - (Optional, ForceNew) The Master image ID, sent as `MasterImageId`. When omitted, the API uses the top-level `image_id` or selects an available base image.
* `boot_disk_size` - (Optional, ForceNew) System disk size in GB, from 40 to 500. When omitted, the API default is 40 GB.
* `boot_disk_type` - (Optional, ForceNew) The type of boot disk. Possible values are: `local_normal` and `local_ssd` for local boot disk, `cloud_ssd` for cloud SSD boot disk,`cloud_rssd` as RDMA-SSD cloud disk. (Default: `cloud_ssd`). The `local_ssd` and `cloud_ssd` are not fully support by all regions as boot disk type, please proceed to UCloud console for more details.
* `data_disk_type` - (Optional, ForceNew) The type of local data disk. Possible values are: `local_normal` and `local_ssd` for local data disk. (Default: `cloud_ssd`). The `local_ssd` is not fully support by all regions as data disk type, please proceed to UCloud console for more details. In addition, the `data_disk_type` must be same as `boot_disk_type` if specified.
* `data_disk_size` - (Optional, ForceNew) The size of local data disk, measured in GB (GigaByte), 20-2000 for local sata disk and 20-1000 for local ssd disk (all the GPU type instances are included). The volume adjustment must be a multiple of 10 GB. In addition, any reduction of data disk size is not supported.
* `min_cpu_platform` - (Optional, ForceNew) Specifies a minimum CPU platform for the VM instance. (Default: `Intel/Auto`). You may refer to [min_cpu_platform](https://docs.ucloud.cn/uhost/introduction/uhost/type_new)
    - The Intel CPU platform:
        - `Intel/Auto` as the Intel CPU platform version will be selected randomly by system;
        - `Intel/IvyBridge` as Intel V2, the version of Intel CPU platform selected by system will be `Intel/IvyBridge` and above;
        - `Intel/Haswell` as Intel V3,  the version of Intel CPU platform selected by system will be `Intel/Haswell` and above;
        - `Intel/Broadwell` as Intel V4, the version of Intel CPU platform selected by system will be `Intel/Broadwell` and above;
        - `Intel/Skylake` as Intel V5, the version of Intel CPU platform selected by system will be `Intel/Skylake` and above;
        - `Intel/Cascadelake` as Intel V6, the version of Intel CPU platform selected by system will be `Intel/Cascadelake`;
        - `Intel/CascadelakeR` as the version of Intel CPU platform, currently can only support by the `os` instance type;
        - `Intel/IceLake`, `Intel/SapphireRapids`, and `Intel/EmeraldRapids` for newer Intel CPU platforms;
    - The AMD CPU platform:
        - `Amd/Auto` as the Amd CPU platform version will be selected randomly by system;
        - `Amd/Epyc2` as the version of Amd CPU platform selected by system will be `Amd/Epyc2` and above;
    - The Ampere CPU platform:
        - `Ampere/Altra` as the version of Ampere CPU platform selected by system will be `Ampere/Altra` and above.

### Master specification example

Use a region-compatible image and zones. For example, the following selects O2 Intel Masters in Singapore:

```hcl
master {
  availability_zones = ["sg-02", "sg-02", "sg-02"]

  machine_type     = "O"
  cpu              = 2
  memory           = 4096
  uhost_family     = "o2i"
  min_cpu_platform = "Intel/EmeraldRapids"
  image_id         = "uimage-1rgr4ndomwqa"

  boot_disk_type = "cloud_rssd"
  boot_disk_size = 40
  data_disk_type = "cloud_rssd"
  data_disk_size = 20
}
```

### Existing configurations and state

Existing configurations using `master.instance_type` continue to work. The version 0 state upgrader preserves the legacy field, resource ID, and historical values without converting them into the independent fields. Keeping the same configuration does not require replacing the cluster when upgrading the provider.

Use either `instance_type` or all three independent specification fields; do not combine them. Use the independent fields for new clusters. Keep `instance_type` in existing configurations: switching an existing cluster to the independent fields changes ForceNew attributes and can require replacement. Review the plan before making that configuration change.

For example, this legacy configuration remains supported:

```hcl
master {
  availability_zones = ["sg-02", "sg-02", "sg-02"]
  instance_type      = "n-basic-2"
}
```

### worker

Each `worker` block creates a group of Workers alongside the cluster. Omit the blocks to create only Masters. Workers share the cluster VPC, subnet, password and billing settings. `worker.image_id` falls back to the top-level `image_id`, not `master.image_id`.

```hcl
# Inside ucloud_uk8s_cluster
worker {
  availability_zone = "sg-02"
  machine_type      = "O"
  cpu               = 2
  memory            = 4096
  count             = 1
  gpu               = 0
  uhost_family      = "o2i"
  min_cpu_platform  = "Intel/EmeraldRapids"
  image_id          = "uimage-1rgr4ndomwqa"
  security_group_id = "291852"

  boot_disk_type = "cloud_rssd"
  boot_disk_size = 40
  data_disk_type = "cloud_rssd"
  data_disk_size = 20
  net_capability = "Ultra"
  max_pods       = 110

  labels = {
    role = "worker"
  }

  taint {
    key    = "dedicated"
    value  = "test"
    effect = "NoSchedule"
  }
}
```

| Field | Requirement / default | API parameter |
| --- | --- | --- |
| `availability_zone` | Required | `Nodes.N.Zone` |
| `machine_type` | Required, uppercase, such as `N` or `O` | `Nodes.N.MachineType` |
| `cpu` | Required, at least 2 cores | `Nodes.N.CPU` |
| `memory` | Required, at least 4096 MB, multiple of 1024 | `Nodes.N.Mem` |
| `count` | Optional, 1–10, default 1 | `Nodes.N.Count` |
| `uhost_family` | Optional, `o1i` or `o2i`; requires `O` and an Intel platform | `Nodes.N.UHostFamily` |
| `min_cpu_platform` | Optional, default `Intel/Auto`; same supported platforms as Master | `Nodes.N.MinimalCpuPlatform` |
| `image_id` | Optional | `Nodes.N.ImageId` |
| `boot_disk_type` | Optional, default `cloud_ssd`; use `cloud_rssd` for `O` machines | `Nodes.N.BootDiskType` |
| `boot_disk_size` | Optional, 40–500 GB, default 40 | `Nodes.N.BootDiskSize` |
| `data_disk_type` | Optional, default `cloud_ssd`; sent when a data disk is requested | `Nodes.N.DataDiskType` |
| `data_disk_size` | Optional, 0 or 20–1000 GB in multiples of 10; default 0 means no data disk | `Nodes.N.DataDiskSize` |
| `gpu` | Optional, non-negative, default 0 | `Nodes.N.GPU` |
| `gpu_type` | Optional; availability depends on machine type and region | `Nodes.N.GpuType` |
| `net_capability` | Optional, `Normal`, `Super`, `Ultra`, `Extreme` | `Nodes.N.NetCapability` |
| `max_pods` | Optional, 1–256, default 110 | `Nodes.N.MaxPods` |
| `security_group_id` | Optional string; the API's firewall ID | `Nodes.N.SecurityGroupId` |
| `labels` | Optional map, at most 5 entries | `Nodes.N.Labels` |
| `taint` | Optional blocks with `key`, optional `value`, and `effect`; at most 5 | `Nodes.N.Taints` |

Disk type values are the same as Master. Taint effects are `NoSchedule`, `PreferNoSchedule`, and `NoExecute`. Multiple Worker blocks may target different zones or specifications; their order determines the API group index.

These blocks record the initial creation configuration in Terraform state. They do not individually track or reconcile subsequently modified, deleted, or added nodes. Adding, removing, reordering, or changing Worker groups requires replacement of the entire cluster, including changes to `count`. For ongoing node management, use `ucloud_uk8s_node_group` and `ucloud_uk8s_node` for separately created nodes. Do not manage the same node through both an initial Worker block and a standalone resource. Cluster deletion continues to use `DelUK8SCluster` and the cluster's `delete_disks_with_instance` setting.

### kube_proxy

* `mode` - (Required, ForceNew) The type of instance, please visit the [instance type table](https://docs.ucloud.cn/terraform/specification/instance)

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 mins) Used when launching the instance (until it reaches the initial `RUNNING` state)
* `update` - (Defaults to 20 mins) Used when updating the arguments of the instance if necessary.
* `delete` - (Defaults to 10 mins) Used when terminating the instance

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource instance.
* `api_server` - The api server endpoint in cluster.
* `external_api_server` - The api server endpoint for external visiting.
* `kubeconfig` - The kubeconfig for accessing the cluster through the internal API server. This value is sensitive.
* `external_kubeconfig` - The kubeconfig for accessing the cluster through the external API server. This value is sensitive and may be empty when external API server access is disabled.
* `pod_cidr` - The CIDR block of pod network.
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `status` - Instance current status. Possible values are `RUNNING`, `CREATEFAILED`, `DELETEFAILED`, `ERROR` and `ABNORMAL`.
