---
subcategory: "Cloud Native"
layout: "ucloud"
page_title: "UCloud: ucloud_uk8s_node"
description: |-
  Provides an UK8S Cluster resource.
---

# ucloud_instance

Provides an UK8S Node resource.

## Example Usage

```hcl
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-uk8s-node"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-uk8s-node"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
  vpc_id       = "${ucloud_vpc.foo.id}"
  subnet_id    = "${ucloud_subnet.foo.id}"
  name         = "tf-acc-uk8s-node-basic"
  service_cidr = "172.16.0.0/16"
  password     = "ucloud_2021"
  charge_type  = "dynamic"

  master {
    availability_zones = [
      "${data.ucloud_zones.default.zones.0.id}",
      "${data.ucloud_zones.default.zones.0.id}",
      "${data.ucloud_zones.default.zones.0.id}",
    ]
    instance_type = "n-basic-2"
  }
}

resource "ucloud_uk8s_node" "foo" {
  cluster_id    = "${ucloud_uk8s_cluster.foo.id}"
  subnet_id     = "${ucloud_subnet.foo.id}"
  password      = "ucloud_2021"
  instance_type = "n-basic-2"
  charge_type   = "dynamic"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  
  count = 2
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where instance is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `cluster_id` - (Required, ForceNew) The ID of uk8s cluster.
* `image_id` - (Required, ForceNew) The ID for the image to use for the instance.
* `password` - (Required, ForceNew) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will auto-generate a password.
* `instance_type` - (Required, ForceNew) The type of instance, please visit the [instance type table](https://docs.ucloud.cn/terraform/specification/instance)
  
---

* `charge_type` - (Optional, ForceNew) The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `boot_disk_type` - (Optional, ForceNew) The type of boot disk. Possible values are: `local_normal` and `local_ssd` for local boot disk, `cloud_ssd` for cloud SSD boot disk,`rssd_data_disk` as RDMA-SSD cloud disk. (Default: `cloud_ssd`). The `local_ssd` and `cloud_ssd` are not fully support by all regions as boot disk type, please proceed to UCloud console for more details.
* `data_disk_type` - (Optional, ForceNew) The type of local data disk. Possible values are: `local_normal` and `local_ssd` for local data disk. (Default: `cloud_ssd`). The `local_ssd` is not fully support by all regions as data disk type, please proceed to UCloud console for more details. In addition, the `data_disk_type` must be same as `boot_disk_type` if specified.
* `data_disk_size` - (Optional, ForceNew) The size of local data disk, measured in GB (GigaByte), 20-2000 for local sata disk and 20-1000 for local ssd disk (all the GPU type instances are included). The volume adjustment must be a multiple of 10 GB. In addition, any reduction of data disk size is not supported. 
* `isolation_group` - (Optional, ForceNew) The ID of the associated isolation group.
* `subnet_id` - (Optional, ForceNew) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. If not defined `vpc_id` and `subnet_id`, the instance will use the default subnet in the current region.
* `user_data` - (Optional, ForceNew) The user data to customize the startup behaviors when launching the instance. You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
* `init_script` - (Optional, ForceNew) The user data to customize the startup behaviors when launching the instance. You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
* `delete_disks_with_instance` - (Optional, ForceNew)  Whether the cloud data disks attached instance should be destroyed on instance termination.

 ~> **NOTE:** We recommend set `delete_disks_with_instance` to `true` means delete cloud data disks attached to instance when instance termination. Otherwise, the cloud data disks will be not managed by the terraform after instance termination.

* `disable_schedule_on_create` - (Optional, ForceNew)  Whether disable any pod scheduling on the node is created.
* `min_cpu_platform` - (Optional) Specifies a minimum CPU platform for the VM instance. (Default: `Intel/Auto`). You may refer to [min_cpu_platform](https://docs.ucloud.cn/uhost/introduction/uhost/type_new)
    - The Intel CPU platform:
        - `Intel/Auto` as the Intel CPU platform version will be selected randomly by system;
        - `Intel/IvyBridge` as Intel V2, the version of Intel CPU platform selected by system will be `Intel/IvyBridge` and above; 
        - `Intel/Haswell` as Intel V3,  the version of Intel CPU platform selected by system will be `Intel/Haswell` and above; 
        - `Intel/Broadwell` as Intel V4, the version of Intel CPU platform selected by system will be `Intel/Broadwell` and above;
        - `Intel/Skylake` as Intel V5, the version of Intel CPU platform selected by system will be `Intel/Skylake` and above; 
        - `Intel/Cascadelake` as Intel V6, the version of Intel CPU platform selected by system will be `Intel/Cascadelake`;
        - `Intel/CascadelakeR` as the version of Intel CPU platform, currently can only support by the `os` instance type;
    - The AMD CPU platform:
        - `Amd/Auto` as the Amd CPU platform version will be selected randomly by system;
        - `Amd/Epyc2` as the version of Amd CPU platform selected by system will be `Amd/Epyc2` and above;
    - The Ampere CPU platform:
        - `Ampere/Altra` as the version of Ampere CPU platform selected by system will be `Ampere/Altra` and above.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 mins) Used when launching the instance (until it reaches the initial `Ready` state)
* `update` - (Defaults to 20 mins) Used when updating the arguments of the instance if necessary  - e.g. when changing `instance_type`
* `delete` - (Defaults to 10 mins) Used when terminating the instance

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource instance.
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `expire_time` - The expiration time for instance, formatted in RFC3339 time string.
* `status` - Instance current status. Possible values are `Initializing`, `Starting`, `Running`, `Stopping`, `Deleting`, `ToBeDeleted`, `Stopped`, `Install Fail` and `Error`.
* `ip_set` - It is a nested type which documented below.

- - -

The attribute (`ip_set`) supports the following:

* `internet_type` - Type of Elastic IP routes. Possible values are: `International` as international BGP IP, `BGP` as china BGP IP and `Private` as private IP.
* `ip` - Elastic IP address.
