---
layout: "ucloud"
page_title: "UCloud: ucloud_instance"
sidebar_current: "docs-ucloud-resource-instance"
description: |-
  Provides an UHost Instance resource.
---

# ucloud_instance

Provides an UHost Instance resource.

~> **Note** The instance will reboot automatically to make the change take effect when update `instance_type`, `root_password`, `boot_disk_size`, `data_disk_size`. In addtion, once the instance complete creation, it takes around 10 mins for boot disk initialization for the running instance, and the updates will only be made to some specific attributes (`root_password`, `boot_disk_size`) if required once the instance initialization completed.

## Example Usage

```hcl
# Query default security group
data "ucloud_security_groups" "default" {
    type = "recommend_web"
}

# Query image
data "ucloud_images" "default" {
  availability_zone = "cn-bj2-04"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

# Create web instance 
resource "ucloud_instance" "web" {
    availability_zone = "cn-bj2-04"
    image_id          = data.ucloud_images.default.images[0].id
    instance_type     = "n-basic-2"
    root_password     = "wA1234567"
    name              = "tf-example-instance"
    tag               = "tf-example"

    # the default Web Security Group that UCloud recommend to users
    security_group = data.ucloud_security_groups.default.security_groups[0].id
}

# Create cloud disk
resource "ucloud_disk" "example" {
    availability_zone = "cn-bj2-04"
    name              = "tf-example-instance"
    disk_size         = 30
}

# Attach cloud disk to instance
resource "ucloud_disk_attachment" "example" {
  availability_zone = "cn-bj2-04"
  disk_id           = ucloud_disk.example.id
  instance_id       = ucloud_instance.web.id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) Availability zone where instance is located. such as: `cn-bj2-02`. You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `image_id` - (Required) The ID for the image to use for the instance.
* `instance_type` - (Required) The type of instance. You may refer to [list of instance type](https://docs.ucloud.cn/compute/terraform/specification/instance)

    ~> **Note**  When it is changed, the instance will reboot automatically to make the change take effect. 
    - The normal type (range of CPU in core: 1-32, range of memory in MB: 1-128, and the number of cores of CPU and memory must be divisible by 2 without a remainder (except single core or memory)):
        - One is normal type defined by UCloud provider: `n-Type-CPU`(eg:`n-highcpu-2`), `Type` can be `highcpu`, `basic`, `standard`, `highmem` which represents the ratio of CPU and memory respectively (1:1, 1:2, 1:4, 1:8), CPU can be the specific number of cores of cpu.
        - One is normal type defined by Customized: `n-customized-CPU-Memory`, the ratio of cpu to memory should be range of 2:1 ~ 1:12 (eg:`n-customized-1-10`). Thereinto, the custmized not be valid when ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8.
    - There is a new outstanding type defined by UCloud provider only valid in availability_zone `cn-bj2-05`: `o-Type-CPU`(eg: `o-standard-4`). `Type` can be `highcpu`, `basic`, `standard`, `highmem` which represents the ratio of CPU and memory respectively (1:1, 1:2, 1:4, 1:8). This type range of CPU in core: 4-64, range of memory in MB: 4-256, the number of cores of CPU and memory must be divisible by 2 without a remainder (except single core or memory). In order to use it, we must set `boot_disk_type` to `cloud_ssd`. In addition, this type needs to be specified to `image_id`, the image type is `base` and the name of which is prefix with "高内核". Furthermore, the disk attached to instance must be `rssd_data_disk` (RDMA-SSD) cloud disk if required. 
* `root_password` - (Optional) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code>. If not specified, terraform will autogenerate a password. 

    ~> **Note** When it is changed, the instance will reboot automatically to make the change take effect.
* `boot_disk_size` - (Optional) The size of the boot disk, measured in GB (GigaByte). Range: 20-100. The value set of disk size must be larger or equal to `20`(default: `20`) for Linux and `40` (default: `40`) for Windows. The responsive time is a bit longer if the value set is larger than default for local boot disk, and further settings may be required on host instance if the value set is larger than default for cloud boot disk. The disk volume adjustment must be a multiple of 10 GB. In addition, any reduction of boot disk size is not supported.

    ~> **Note** When it is changed, the instance will reboot automatically to make the change take effect and need to [go to the instance for configuration](https://docs.ucloud.cn/compute/uhost/guide/disk). 
* `boot_disk_type` - (Optional) The type of boot disk. Possible values are: `local_normal` and `local_ssd` for local boot disk, `cloud_normal` and `cloud_ssd` for cloud boot disk. (Default: `local_normal`). The `local_ssd`, `cloud_normal` and `cloud_ssd` are not fully support by all regions as boot disk type, please proceed to UCloud console for more details.
* `data_disk_type` - (Optional) The type of local data disk. Possible values are: `local_normal` and `local_ssd` for local data disk. (Default: `local_normal`). The `local_ssd` is not supported in all regions as data disk type, please proceed to UCloud console for more details.
* `data_disk_size` - (Optional) The size of local data disk, measured in GB (GigaByte), range: 0-8000 (Default: `20`), 0-8000 for cloud disk, 0-2000 for local sata disk and 100-1000 for local ssd disk (all the GPU type instances are included). The volume adjustment must be a multiple of 10 GB. In addition, any reduction of data disk size is not supported. 

    ~> **Note** When it is changed, the instance will reboot automatically to make the change take effect and need to [go to the instance for configuration](https://docs.ucloud.cn/compute/uhost/guide/disk). 
* `charge_type` - (Optional) The charge type of instance, possible values are: `year`, `month` and `dynamic` as pay by hour (specific permission required). (Default: `month`).
* `duration` - (Optional) The duration that you will buy the instance (Default: `1`). The value is `0` when pay by month and the instance will be vaild till the last day of that month. It is not required when `dynamic` (pay by hour).
* `name` - (Optional) The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will autogenerate a name beginning with `tf-instance`.
* `remark` - (Optional) The remarks of instance. (Default: `""`).
* `security_group` - (Optional) The ID of the associated security group.
* `subnet_id` - (Optional) The ID of subnet. If defined `vpc_id`, the `subnet_id` is Required. If not defined `vpc_id` and `subnet_id`, the instance will use the default subnet in the current region.
* `tag` - (Optional) A tag assigned to instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: `Default`).
* `vpc_id` - (Optional) The ID of VPC linked to the instance. If not defined `vpc_id`, the instance will use the default VPC in the current region.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `auto_renew` - Whether to renew an instance automatically or not.
* `cpu` - The number of cores of virtual CPU, measureed in core.
* `memory` - The size of memory, measured in MB (Megabyte).
* `create_time` - The time of creation for instance, formatted in RFC3339 time string.
* `expire_time` - The expiration time for instance, formatted in RFC3339 time string.
* `status` - Instance current status. Possible values are `Initializing`, `Starting`, `Running`, `Stopping`, `Stopped`, `Install Fail`, `ResizeFail` and `Rebooting`.
* `private_ip` - The private IP address assigned to the instance.
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

## Import

Instance can be imported using the `id`, e.g.

```
$ terraform import ucloud_instance.example uhost-abcdefg
```