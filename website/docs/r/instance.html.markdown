---
layout: "ucloud"
page_title: "UCloud: ucloud_instance"
sidebar_current: "docs-ucloud-resource-instance"
description: |-
  Provides an UHost Instance resource.
---

# ucloud_instance

Provides an UHost Instance resource.

## Example Usage

```hcl
resource "ucloud_security_group" "default" {
    name = "tf-example-instance"
    tag  = "tf-example"

    # HTTP access from LAN
    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }

    # HTTPS access from LAN
    rules {
        port_range = "443"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
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
    availability_zone = "cn-sh2-02"
    image_id          = "uimage-of3pac"
    instance_type     = "n-standard-1"

    # use cloud disk as data disk
    data_disk_size     = 50
    data_disk_category = "Disk"
    root_password      = "wA1234567"

    # we will put all the instances into same vpc and subnet,
    # so they can communicate with each other.
    vpc_id    = "${ucloud_vpc.default.id}"
    subnet_id = "${ucloud_subnet.default.id}"

    # this ecurity group to allow HTTP and HTTPS access
    security_group = "${ucloud_security_group.default.id}"
}

```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) Availability zone where instance is located. such as: "cn-bj-01". You may refer to [list of availability zone](https://docs.ucloud.cn/api/summary/regionlist)
* `image_id` - (Required) The ID for the image to use for the instance.
* `root_password` - (Required) The password for the instance, should have between 8-30 characters.It must contain least 3 items of Capital letters, small letter, numbers and special characters. The special characters incloud <code>`()~!@#$%^&*-+=_|{}\[]:;'<>,.?/</code> When it is changed, the instance will reboot to make the change take effect.
* `instance_type` - (Required) The type of instance.There are two types, one is Customized: "n-customized-CPU-Memory", eg."n-customized-1-3",the other is Standard: "n-Type-CPU", eg."n-highcpu-2". Thereinto,"Type" can be "highcpu", "basic", "standard", "highmem" represent the ratio of CPU and Memory respectively, 1:1, 1:2, 1:4, 1:8. In addition, CPU range from 1 to 32 ,Memory range from 1 to 128. When it is changed, the instance will reboot to make the change take effect.
* `boot_disk_size` - (Optional) [will be invalid soon, not recommend to call] Size of the system disk, measured in GB (Giga byte). when the instance is created, the system disk fixed in size, 20GB for Linux system, 40GB for Windows system. when the instance is updated, the system disk range from 20GB to 100 GB, the volume adjustment must be a multiple of 10 GB. When it is changed, the instance will reboot to make the change take effect.
* `data_disk_category` - (Optional) [will be invalid soon, not recommend to call] The type of disk, the defination of disk type for both system disk and data disk. Possible values are: "LocalDisk" and "Disk" as cloud disk, the default is "LocalDisk". The "Disk" is not supported in all regions as disk type, please proceed to UCloud console for more details.
* `data_disk_size` - (Optional) [will be invalid soon, not recommend to call] Size of data disk, measured in GB (Giga byte), range from 0 to 8000 GB, the volume adjustment must be a multiple of 10 GB, default is 20 GB. Volume is from 0 to 8000GB as cloud disk, from 0 to 2000GB as local sata disk and from 100 to 1000GB as local ssd disk (all the GPU type instances are included). When it is changed, the instance will reboot to make the change take effect.
* `instance_charge_type` - (Optional) The charge type of instance, possible values are: "Year", "Month" and "Dynamic" as pay by hour (specific permission required). the dafault is "Month".
* `instance_duration` - (Optional) The duration that you will buy the resource, the default value is "1". It is not required when "Dynamic" (pay by hour), the value is "0" when pay by month and the instance will be vaild till the last day of that month.
* `name` - (Optional) The name of instance, the default is "Instance", should have 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'.
* `remark` - (Optional) The remarks of instance,the default value is "".
* `security_group` - (Optional) The ID of the associated security group.
* `subnet_id` - (Optional) The ID of subnet.
* `tag` - (Optional) A mapping of tags to assign to the instance. The default value is "Default" (means no tag assigned).
* `vpc_id` - (Optional) The ID of VPC linked to the instances.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `auto_renew` - Whether to renew an ECS instance automatically or not. Passible values are "Yes" as enabling auto renewal and "No" as disabling auto renewal.
* `cpu` - The number of cores of virtual CPU, measureed in core.
* `memory` - The size of memory, measured in MB (Megabyte).
* `create_time` - The time of creation for instance.
* `expire_time` - The expiration time for instance.
* `status` - Instance current status. Possible values are "Initializing", "starting", "Running", "Stopping", "Stopped", "Install Fail" and "Rebooting".
* `ip_set` - ip_set is a nested type. ip_set documented below.
* `disk_set` - disk_set is a nested type. disk_set documented below.

The attribute (`disk_set`) support the following:

* `disk_id` - The ID of disk.
* `size` - The size of disk，measured in GB (Gigabyte).
* `type` - The type of disk. Possible values are "Boot" as system disk, "Data" as data disk and "Disk" as cloud disk.

The attribute (`ip_set`) support the following:

* `bandwidth` - The size of bandwidth for the corresponding IP address, measured in Mbps (Mega bit per second), This atttibute is not exported if intranet IPs.
* `internet_type` - The attached EIP route. Possible values are "Internaltionl" , "Bgp" and "Private" as intranet.
* `ip` - IP address.
