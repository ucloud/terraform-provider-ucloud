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
* `vpc_id` - (Required, ForceNew) The ID of VPC linked to the instance. If not defined `vpc_id`, the instance will use the default VPC in the current region.
* `subnet_id` - (Required, ForceNew) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. If not defined `vpc_id` and `subnet_id`, the instance will use the default subnet in the current region.
* `password` - (Required) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will auto-generate a password.

---

* `name` - (Optional) The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will auto-generate a name beginning with `tf-instance`.
* `user_data` - (Optional, ForceNew) Plain-text user data, up to 16 KiB. The provider base64-encodes it and sends `UserData` when creating the cluster. It customizes startup behaviors when launching the instance. You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
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

### Additional cluster creation options

All options below are optional and ForceNew. They are omitted from the creation request when not configured, leaving API defaults in effect. These are creation settings retained in state; the provider does not reconcile out-of-band changes to them. Adding, changing or removing them on an existing cluster replaces the cluster. Review the Terraform plan before applying.

| Terraform field | API parameter | Description |
| --- | --- | --- |
| `cluster_domain` | `ClusterDomain` | Cluster DNS domain. |
| `tag` | `Tag` | UCloud business group. |
| `lb_class` | `LbClass` | Master load balancer type: `ulb` or `nlb`; API default is `ulb`. |
| `forward_src_ip_method` | `ForwardSrcIPMethod` | `Toa` requires `lb_class = "nlb"`. Empty or omitted disables source IP forwarding. |
| `kms_plugin_key_id` | `KmsPluginKeyId` | KMS key for the cluster encryption plugin. |
| `kms_plugin_resources` | `KmsPluginResource.N` | Set of resources to encrypt, such as `["secrets"]`; requires `kms_plugin_key_id`. |
| `user_labels` | `UserLabels.N.Key`, `UserLabels.N.Value` | Map of UCloud resource labels. These are separate from Kubernetes `worker.labels`. |

For example, inside `ucloud_uk8s_cluster`:

```hcl
cluster_domain        = "cluster.local"
tag                   = "platform"
lb_class              = "nlb"
forward_src_ip_method = "Toa"
kms_plugin_key_id     = "your-kms-key-id"
kms_plugin_resources  = ["secrets"]
user_data             = "#!/bin/sh\necho preparing-node"

user_labels = {
  environment = "production"
  owner       = "platform"
}
```

The obsolete `MasterIsolationGroup` parameter is not exposed; UK8S manages Master isolation automatically.

### master

The `master` supports the following:

* `availability_zones` - (Required, ForceNew) Availability zone list where instance is located. such as: `["cn-bj2-02", "cn-bj2-03", "cn-bj2-05"]`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `instance_type` - (Optional, Deprecated, ForceNew) The legacy specification, such as `n-basic-2`. Existing configurations remain supported with the original validation rules. Conflicts with `machine_type`, `cpu`, and `memory`.
* `machine_type` - (Optional, ForceNew) The uppercase machine type, such as `N` or `O`. Required together with `cpu` and `memory` when `instance_type` is omitted.
* `cpu` - (Optional, ForceNew) Number of CPU cores, at least 2. Required with `machine_type` and `memory` when `instance_type` is omitted. Available specifications depend on the machine type and region.
* `memory` - (Optional, ForceNew) Memory in MB, at least 4096 and a multiple of 1024. Required with `machine_type` and `cpu` when `instance_type` is omitted. For 4 GB, use `4096`.
* `uhost_family` - (Optional, ForceNew) The host family, such as `o1a`, `o1i`, or `o2i`, sent as `MasterUHostFamily`. Lowercase alphanumeric family names are accepted; availability and compatibility are checked by the API. Existing `o1i`/`o2i` validation still requires `machine_type = "O"` and an Intel CPU platform. Use a matching CPU platform (for example `Amd/Auto` for `o1a`). Outstanding machines require `boot_disk_type = "cloud_rssd"`.
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

### Master security groups and disk encryption

* `master.data_disk_kms_key_id` - (Optional, ForceNew) KMS key ID sent as `MasterDataDiskKmsKeyId`, for the Master data disk.
* `master.security_group` - (Optional, ForceNew) Repeatable security group binding. Each binding has required `master_index` (0, 1, or 2, corresponding to `availability_zones`), required `id`, optional `priority` (1–5), and optional `name`. Up to five distinct groups can be bound to each Master. The same group can be used on different Masters. This maps to `Master.N.SecGroupId.N.{Id,Priority,Name}`.

For example, within the existing `master` block:

```hcl
# Alongside the existing zone, machine and disk settings:
data_disk_kms_key_id = "your-master-disk-kms-key-id"

security_group {
  master_index = 0
  id           = "your-security-group-id"
  priority     = 1
  name         = "master-access"
}
# Add bindings for master_index 1 and 2 as needed.
```

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
| `uhost_family` | Optional host family, such as `o1a`, `o1i`, or `o2i`; same validation as Master | `Nodes.N.UHostFamily` |
| `min_cpu_platform` | Optional, default `Intel/Auto`; same supported platforms as Master | `Nodes.N.MinimalCpuPlatform` |
| `image_id` | Optional | `Nodes.N.ImageId` |
| `boot_disk_type` | Optional, default `cloud_ssd`; use `cloud_rssd` for `O` machines | `Nodes.N.BootDiskType` |
| `boot_disk_size` | Optional, 40–500 GB, default 40 | `Nodes.N.BootDiskSize` |
| `data_disk_type` | Optional, default `cloud_ssd`; sent when a data disk is requested | `Nodes.N.DataDiskType` |
| `data_disk_size` | Optional, 0 or 20–1000 GB in multiples of 10; default 0 means no data disk | `Nodes.N.DataDiskSize` |
| `gpu` | Optional, non-negative, default 0 | `Nodes.N.GPU` |
| `gpu_type` | Optional; availability depends on machine type and region | `Nodes.N.GpuType` |
| `max_pods` | Optional, 1–256, default 110 | `Nodes.N.MaxPods` |
| `security_group_id` | Optional string; the API's firewall ID | `Nodes.N.SecurityGroupId` |
| `labels` | Optional map, at most 5 entries | `Nodes.N.Labels` |
| `taint` | Optional blocks with `key`, optional `value`, and `effect`; at most 5 | `Nodes.N.Taints` |

Disk type values are the same as Master. Taint effects are `NoSchedule`, `PreferNoSchedule`, and `NoExecute`. Multiple Worker blocks may target different zones or specifications; their order determines the API group index.

These blocks record the initial creation configuration in Terraform state. They do not individually track or reconcile subsequently modified, deleted, or added nodes. Adding, removing, reordering, or changing Worker groups requires replacement of the entire cluster, including changes to `count`. For ongoing node management, use `ucloud_uk8s_node_group` and `ucloud_uk8s_node` for separately created nodes. Do not manage the same node through both an initial Worker block and a standalone resource. Cluster deletion continues to use `DelUK8SCluster` and the cluster's `delete_disks_with_instance` setting.

### Additional Worker creation options

All options below belong inside a `worker` block, are optional and ForceNew, and follow the initial Worker lifecycle described above.

| Field | API parameter | Description |
| --- | --- | --- |
| `isolation_group` | `Nodes.N.IsolationGroup` | Worker isolation group ID. The API limits each isolation group to eight nodes, including existing nodes. |
| `name_prefix` | `Nodes.N.NamePrefix` | Hostname prefix; the resulting hostname is `{NamePrefix}-{NodeIP}`. |
| `security_mode` | `Nodes.N.SecurityMode` | `Firewall` or `SecGroup`; omitted uses the API default `Firewall`. |
| `security_group` | `Nodes.N.SecGroupId.N.*` | Up to five bindings, each with required `id`, optional `priority` (1–5), and optional `name`. Requires `security_mode = "SecGroup"`. |
| `uni_feature` | `Nodes.N.UNIFeature` | String `"true"` or `"false"`; omitted uses the API default. Enabling requires the relevant UCloud permission. |
| `data_disk_kms_key_id` | `Nodes.N.DataDiskKmsKeyId` | KMS key ID for the Worker data disk. |
| `network_interface` | `Nodes.N.NetworkInterface.N.*` | Repeatable network interface configuration, each containing one `eip` block. |
| `kubelet_configuration` | `Nodes.N.KubeletConfiguration.ContainerLogMaxFiles` | Map supporting only `ContainerLogMaxFiles`, as an integer string of at least 2. |

`security_group_id` remains the legacy firewall ID. It cannot be combined with `security_mode = "SecGroup"`. This differs from the `security_group` blocks used for the newer security groups.

Each `network_interface.eip` block supports:

| Field | Requirement / values | API field |
| --- | --- | --- |
| `operator_name` | Required: `Bgp` or `International`, according to the region | `OperatorName` |
| `pay_mode` | Optional: `Bandwidth` (default), `Traffic`, `ShareBandwidth`, `Free` | `PayMode` |
| `bandwidth` | Positive Mbps required outside shared bandwidth mode; `Traffic` 1–300, `Bandwidth` 1–800 | `Bandwidth` |
| `share_bandwidth_id` | Required for `ShareBandwidth`; only valid with that mode | `ShareBandwidthId` |
| `coupon_id` | Optional EIP coupon ID | `CouponId` |

For example, within an existing `worker` block:

```hcl
isolation_group      = "your-isolation-group-id"
name_prefix          = "worker"
security_mode        = "SecGroup"
uni_feature          = "true"
data_disk_kms_key_id = "your-worker-disk-kms-key-id"

security_group {
  id       = "your-security-group-id"
  priority = 1
  name     = "worker-access"
}

network_interface {
  eip {
    operator_name = "International"
    pay_mode      = "Bandwidth"
    bandwidth     = 10
  }
}

kubelet_configuration = {
  ContainerLogMaxFiles = "5"
}
```

Only `ContainerLogMaxFiles` is supported in `kubelet_configuration`. It sets the maximum number of container log files and must be an integer string of at least 2. Omitting the map leaves the API default in effect. Unknown keys are rejected during planning; changing this creation setting replaces the cluster.

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
* `cni_mode` - The CNI network mode reported by the cluster.
* `api_server` - The api server endpoint in cluster.
* `external_api_server` - The api server endpoint for external visiting.
* `kubeconfig` - The kubeconfig for accessing the cluster through the internal API server. This value is sensitive.
* `external_kubeconfig` - The kubeconfig for accessing the cluster through the external API server. This value is sensitive and may be empty when external API server access is disabled.
* `pod_cidr` - The CIDR block of pod network.
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `status` - Instance current status. Possible values are `RUNNING`, `CREATEFAILED`, `DELETEFAILED`, `ERROR` and `ABNORMAL`.
