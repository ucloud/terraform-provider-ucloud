---
subcategory: "UK8S"
layout: "ucloud"
page_title: "UCloud: ucloud_uk8s_cluster"
description: |-
  Provides an UK8S Cluster resource.
---

# ucloud_instance

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
  vpc_id     = ucloud_vpc.foo.id
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
  vpc_id               = ucloud_vpc.foo.id
  subnet_id            = ucloud_subnet.foo.id
  name                 = "tf-acc-uk8s-cluster-basic"
  service_cidr         = "172.16.0.0/16"
  password             = var.password
  charge_type          = "dynamic"
  master_instance_type = "n-basic-2"
  master {
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
  master {
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
  master {
    availability_zone = data.ucloud_zones.default.zones.0.id
  }

  nodes {
    instance_type     = "n-basic-2"
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
  nodes {
    instance_type     = "n-basic-4"
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, ForceNew) Availability zone where instance is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `image_id` - (Required) The ID for the image to use for the instance.
* `instance_type` - (Required) The type of instance, please visit the [instance type table](https://docs.ucloud.cn/terraform/specification/instance)

    ~> **Note** If you want to update this value, you must set `allow_stopping_for_update`to `true`.

- - -

* `allow_stopping_for_update` - (Optional) If you try to update some properties which requires stopping the instance, you must set `allow_stopping_for_update` to `true` in your config to allows Terraform to stop the instance to update its properties like `instance_type`, `root_password`, `boot_disk_size`, `data_disk_size`.
 
* `root_password` - (Optional) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will auto-generate a password. 

    ~> **Note** If you want to update this value, you must set `allow_stopping_for_update`to `true`.
* `boot_disk_size` - (Optional) The size of the boot disk, measured in GB (GigaByte). Range: 20-500. The value set of disk size must be larger or equal to `20`(default: `20`) for Linux and `40` (default: `40`) for Windows. The responsive time is a bit longer if the value set is larger than default for local boot disk, and further settings may be required on host instance if the value set is larger than default for cloud boot disk. The disk volume adjustment must be a multiple of 10 GB. In addition, any reduction of boot disk size is not supported.

    ~> **Note** If you want to update this value, you must set `allow_stopping_for_update`to `true`. In addition, when it is changed, you need to [go to the instance for configuration](https://docs.ucloud.cn/compute/uhost/guide/disk). 
* `boot_disk_type` - (Optional, ForceNew) The type of boot disk. Possible values are: `local_normal` and `local_ssd` for local boot disk, `cloud_ssd` for cloud SSD boot disk,`rssd_data_disk` as RDMA-SSD cloud disk. (Default: `local_normal`). The `local_ssd` and `cloud_ssd` are not fully support by all regions as boot disk type, please proceed to UCloud console for more details.
* `data_disk_type` - (Optional, ForceNew) The type of local data disk. Possible values are: `local_normal` and `local_ssd` for local data disk. (Default: `local_normal`). The `local_ssd` is not fully support by all regions as data disk type, please proceed to UCloud console for more details. In addition, the `data_disk_type` must be same as `boot_disk_type` if specified.
* `data_disk_size` - (Optional) The size of local data disk, measured in GB (GigaByte), 20-2000 for local sata disk and 20-1000 for local ssd disk (all the GPU type instances are included). The volume adjustment must be a multiple of 10 GB. In addition, any reduction of data disk size is not supported. 

    ~> **Note** If you want to update this value, you must set `allow_stopping_for_update`to `true`. In addition, when it is changed, you need to [go to the instance for configuration](https://docs.ucloud.cn/compute/uhost/guide/disk). 
* `charge_type` - (Optional, ForceNew) The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional, ForceNew) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be valid till the last day of that month. It is not required when `dynamic` (pay by hour).
* `name` - (Optional) The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will auto-generate a name beginning with `tf-instance`.
* `remark` - (Optional) The remarks of instance. (Default: `""`).
* `security_group` - (Optional) The ID of the associated security group.
* `vpc_id` - (Optional, ForceNew) The ID of VPC linked to the instance. If not defined `vpc_id`, the instance will use the default VPC in the current region.
* `subnet_id` - (Optional, ForceNew) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. If not defined `vpc_id` and `subnet_id`, the instance will use the default subnet in the current region.
* `tag` - (Optional) A tag assigned to instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `isolation_group` - (Optional, ForceNew) The ID of the associated isolation group.
* `private_ip` - (Optional, ForceNew) The private IP address assigned to the instance.
* `user_data` - (Optional, ForceNew) The user data to customize the startup behaviors when launching the instance. You may refer to [user_data_document](https://docs.ucloud.cn/uhost/guide/metadata/userdata)
* `data_disks` - (Optional, ForceNew) Additional cloud data disks to attach to the instance. `data_disks` configurations only apply on resource creation. The count of `data_disks` can only be one. See [data_disks](#data_disks) below for details on attributes. When set `data_disks`, the argument `delete_disks_with_instance` must bet set.
* `delete_disks_with_instance` - (Optional, ForceNew, Required when set `data_disks`)  Whether the cloud data disks attached instance should be destroyed on instance termination.

 ~> **NOTE:** We recommend set `delete_disks_with_instance` to `true` means delete cloud data disks attached to instance when instance termination. Otherwise, the cloud data disks will be not managed by the terraform after instance termination.
 
 * `network_interface` - (Optional, ForceNew) Additional network interface eips to attach to the instance. `network_interface` configurations only apply on resource creation. The count of `network_interface` can only be one. See [network_interface](#network_interface) below for details on attributes. When set `network_interface`, the argument `delete_eips_with_instance` must bet set.
 * `delete_eips_with_instance` - (Optional, ForceNew, Required when set `network_interface`)  Whether the network interface eips associated instance should be destroyed on instance termination.
 
* `min_cpu_platform` - (Optional) Specifies a minimum CPU platform for the the VM instance. (Default: `Intel/Auto`). You may refer to [min_cpu_platform](https://docs.ucloud.cn/uhost/introduction/uhost/type_new)
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
        
### data_disks

The `data_disks` supports the following:

* `size` - (Required, ForceNew) The size of the cloud data disk, range 20-8000, measured in GB (GigaByte).
* `type` - (Required, ForceNew) The type of the cloud data disk. Possible values are: `cloud_normal` for cloud disk, `cloud_ssd` for SSD cloud disk, `rssd_data_disk` as RDMA-SSD cloud disk. 

~> **NOTE:** Currently, changes to the `data_disks` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes and attachments of a cloud data disk to an instance, use the `ucloud_disk` and `ucloud_disk_attachment` resources instead. `data_disks` cannot be mixed with external `ucloud_disk` and `ucloud_disk_attachment` resources for a given instance.  Recently, we recommend using `data_disks` to allocate cloud data disk attached to instance.

### network_interface

The `network_interface` supports the following:

* `eip_bandwidth` - (Required, ForceNew) Maximum bandwidth to the elastic public network, measured in Mbps (Mega bit per second). The ranges for bandwidth are: 1-200 for pay by traffic, 1-800 for pay by bandwidth. 
* `eip_internet_type` - (Required, ForceNew) Type of Elastic IP routes. Possible values are: `international` as international BGP IP and `bgp` as china mainland BGP IP.
* `eip_charge_mode` - (Required, ForceNew) Elastic IP charge mode. Possible values are: `traffic` as pay by traffic, `bandwidth` as pay by bandwidth mode.

~> **NOTE:** Currently, changes to the `network_interface` configuration of _existing_ resources cannot be automatically detected by Terraform. To manage changes and associations of an Elastic IP to an instance, use the `ucloud_eip` and `ucloud_eip_association` resources instead. `network_interface` cannot be mixed with external `ucloud_eip` and `ucloud_eip_association` resources for a given instance.  Recently, we recommend using `network_interface` to allocate cloud data disk attached to instance.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 mins) Used when launching the instance (until it reaches the initial `Running` state)
* `update` - (Defaults to 20 mins) Used when updating the arguments of the instance if necessary  - e.g. when changing `instance_type`
* `delete` - (Defaults to 10 mins) Used when terminating the instance

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resource instance.
* `auto_renew` - Whether to renew an instance automatically or not.
* `cpu` - The number of cores of virtual CPU, measured in core.
* `memory` - The size of memory, measured in GB(Gigabyte).
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `expire_time` - The expiration time for instance, formatted in RFC3339 time string.
* `status` - Instance current status. Possible values are `Initializing`, `Starting`, `Running`, `Stopping`, `Stopped`, `Install Fail`, `ResizeFail` and `Rebooting`.
* `ip_set` - It is a nested type which documented below.
* `disk_set` - It is a nested type which documented below.

- - -

The attribute (`disk_set`) supports the following:

* `id` - The ID of disk.
* `size` - The size of disk, measured in GB (Gigabyte).
* `type` - The type of disk.
* `is_boot` - Specifies whether boot disk or not.

The attribute (`ip_set`) supports the following:

* `internet_type` - Type of Elastic IP routes. Possible values are: `International` as international BGP IP, `BGP` as china BGP IP and `Private` as private IP.
* `ip` - Elastic IP address.

## Import

Instance can be imported using the `id`, e.g.

```
$ terraform import ucloud_instance.example uhost-abcdefg
```