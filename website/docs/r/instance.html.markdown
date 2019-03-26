---
layout: "ucloud"
page_title: "UCloud: ucloud_instance"
sidebar_current: "docs-ucloud-resource-instance"
description: |-
  Provides an UHost Instance resource.
---

# ucloud_instance

Provides an UHost Instance resource.

~> **Note** The instance will reboot automatically to make the change take effect when update `instance_type`, `root_password`, `boot_disk_size`, `data_disk_size`.

## Example Usage

```hcl
resource "ucloud_security_group" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"

    # http access from LAN
    rules {
        port_range = "80"
        protocol   = "tcp"
        cidr_block = "192.168.0.0/16"
        policy     = "accept"
    }

    # https access from LAN
    rules {
        port_range = "443"
        protocol   = "tcp"
        cidr_block = "192.168.0.0/16"
        policy     = "accept"
    }
}

resource "ucloud_vpc" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"

    # vpc network
    cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"

    # subnet's network must be contained by vpc network
    # and a subnet must have least 8 ip addresses in it (netmask < 30).
    cidr_block = "192.168.1.0/24"
    vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_instance" "web" {
    name              = "tf-example-instance"
    tag               = "tf-example"
    availability_zone = "cn-bj2-02"
    image_id          = "uimage-of3pac"
    instance_type     = "n-standard-1"

    # use cloud disk as data disk
    data_disk_size     = 50
    data_disk_type     = "local_normal"
    root_password      = "wA1234567"

    # we will put all the instances into same vpc and subnet,
    # so they can communicate with each other.
    vpc_id    = "${ucloud_vpc.default.id}"
    subnet_id = "${ucloud_subnet.default.id}"

    # this security group to allow http and https access
    security_group = "${ucloud_security_group.default.id}"
}

```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) Availability zone where instance is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `image_id` - (Required) The ID for the image to use for the instance.
* `instance_type` - (Required) The type of instance. There are two types, one is defined by UCloud provider: `n-Type-CPU`(eg:`n-highcpu-2`), thereinto, `Type` can be `highcpu`, `basic`, `standard`, `highmem` which represent the ratio of CPU and memory respectively (1:1, 1:2, 1:4, 1:8). The other is defined Customized: `n-customized-CPU-Memory`(eg:`n-customized-1-3`). Be attention, if the type can be defined by `n-Type-CPU`, another type will not be allowed. In addition, range of CPU in core: 1-32, range of memory in MB: 1-256. When it is changed, the instance will reboot automatically to make the change take effect.
* `root_password` - (Optional) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will autogenerate a password. Note: When it is changed, the instance will reboot automatically to make the change take effect.
* `boot_disk_size` - (Optional) The size of the boot disk, measured in GB (GigaByte). Range: 20-100. The value set of disk size must be larger or equal to `20`(default: `20`) for Linux and `40` (default: `40`) for Windows. The responsive time is a bit longer if the value set is larger than default for local boot disk, and further settings may be required on host instance if the value set is larger than default for cloud boot disk. The disk volume adjustment must be a multiple of 10 GB. When it is changed, the instance will reboot automatically to make the change take effect and need to [go to the instance for configuration](https://docs.ucloud.cn/compute/uhost/guide/disk). In addition, any reduction of boot disk size is not supported.
* `boot_disk_type` - (Optional) The type of boot disk. Possible values are: `local_normal` and `local_ssd` for local boot disk, `cloud_normal` and `cloud_ssd` for cloud boot disk. (Default: `local_normal`). The `local_ssd`, `cloud_normal` and `cloud_ssd` are not fully support by all regions as boot disk type, please proceed to UCloud console for more details.
* `data_disk_type` - (Optional) The type of local data disk. Possible values are: `local_normal` and `local_ssd` for local data disk. (Default: `local_normal`). The `local_ssd` is not supported in all regions as data disk type, please proceed to UCloud console for more details.
* `data_disk_size` - (Optional) The size of data disk, measured in GB (GigaByte), range: 0-8000 (Default: `20`), 0-8000 for cloud disk, 0-2000 for local sata disk and 100-1000 for local ssd disk (all the GPU type instances are included). The volume adjustment must be a multiple of 10 GB. When it is changed, the instance will reboot automatically to make the change take effect and need to [go to the instance for configuration](https://docs.ucloud.cn/compute/uhost/guide/disk). In addition, any reduction of data disk size is not supported.
* `charge_type` - (Optional) The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be vaild till the last day of that month. It is not required when `dynamic` (pay by hour).
* `name` - (Optional) The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will autogenerate a name beginning with `tf-instance`.
* `remark` - (Optional) The remarks of instance. (Default: `""`).
* `security_group` - (Optional) The ID of the associated security group.
* `subnet_id` - (Optional) The ID of subnet.
* `tag` - (Optional) A mapping of tags to assign to VPC, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `vpc_id` - (Optional) The ID of VPC linked to the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `auto_renew` - Whether to renew an instance automatically or not.
* `cpu` - The number of cores of virtual CPU, measureed in core.
* `memory` - The size of memory, measured in MB (Megabyte).
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `expire_time` - The expiration time for instance, formatted in RFC3339 time string.
* `status` - Instance current status. Possible values are `Initializing`, `Starting`, `Running`, `Stopping`, `Stopped`, `InstallFail`, `ResizeFail` and `Rebooting`.
* `ip_set` - It is a nested type which documented below.
* `disk_set` - It is a nested type which documented below.

The attribute (`disk_set`) supports the following:

* `id` - The ID of disk.
* `size` - The size of disk, measured in GB (Gigabyte).
* `type` - The type of disk.
* `is_boot` - Specifies whether boot disk or not.

The attribute (`ip_set`) supports the following:

* `internet_type` - Type of Elastic IP routes. Possible values are: `International` as internaltional BGP IP, `BGP` as china BGP IP and `Private` as private IP.
* `ip` - Elastic IP address.