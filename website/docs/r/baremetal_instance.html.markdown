---
subcategory: "Bare Metail (UPhost)"
layout: "ucloud"
page_title: "UCloud: ucloud_baremetal_instance"
sidebar_current: "docs-ucloud-resource-baremetal-instance"
description: |-
  Provides a UCloud Bare Metal instance resource.

---

# ucloud_baremetal_instance

The `ucloud_baremetal_instance` resource provides a UCloud Bare Metal instance. This can be used to create, modify, and delete Bare Metal instances.

## Example Usage
```hcl
resource "ucloud_baremetal_instance" "example" {
  availability_zone = "cn-bj2-02"
  image_id          = "pimg-cs-aqxttl"
  root_password     = "test123456"
  network_interface {
    eip_bandwidth = 10
    eip_charge_mode = "traffic"
    eip_internet_type = "bgp"
  }
  tag = "Default"
  instance_type     = "Base-SSD-V5"
  name              = "UPHost"
  raid_type         = "no_raid" 
  charge_type = "day"
  vpc_id = "uvnet-xxx"
  subnet_id = "subnet-yyy"
  security_group = "firewall-zzz"
}
```

## Argument Reference

The following arguments are supported:

- `availability_zone` - (Required, ForceNew) The availability zone where the instance is created.
- `instance_type` - (Required, ForceNew) The instance type of expected bare metal instance.
- `image_id` - (Required) The ID of the image used to launch the instance. It can be got from ucloud_baremetal_images datasource according to instance_type.
- `allow_stopping_for_update` - (Optional) If you try to update some properties which requires stopping the instance, you must set allow_stopping_for_update to true in your config to allows Terraform to stop the instance to update its properties like root_password,
- `allow_stopping_for_resizing` - (Optional) Allow stopping the instance when the boot disk size needs to be resized.
- `delete_disks_with_instance` - (Optional, ForceNew) Whether the cloud data disks attached should be destroyed on instance termination.
- `delete_eips_with_instance` - (Optional, ForceNew) Whether the EIP associated should be destroyed on instance termination.
- `root_password` - (Optional) The password for the instance, which contains 8-30 characters, and at least 2 items of capital letters, lower case letters, numbers and special characters. The special characters include `()~!@#$%^&*-+=_|{}\[]:;'<>,.?/. If not specified, terraform will auto-generate a password.
- `boot_disk_size` - (Optional) The size of the boot disk, measured in GB (GigaByte). Range: 20-500. The value set of disk size must be larger or equal to 20(default: 20) for Linux and 40 (default: 40) for Windows. The responsive time is a bit longer if the value set is larger than default for local boot disk, and further settings may be required on host instance if the value set is larger than default for cloud boot disk. The disk volume adjustment must be a multiple of 10 GB. In addition, any reduction of boot disk size is not supported. Only cloud disk type instance can have a custom boot disk.
- `boot_disk_type` - (Optional, ForceNew) The type of boot disk. Now only cloud_rssd is supported. Only cloud disk type instance can have a custom boot disk.
- `charge_type` - (Optional, ForceNew) The charge type of instance, possible values are: year, month and day. (Default: month).
- `duration` - (Optional, ForceNew) The duration that you will buy the instance (Default: 1). The value is 0 when pay by month and the instance will be valid till the last day of that month.
- `name` - (Optional) The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'. If not specified, terraform will auto-generate a name beginning with tf-instance.
- `remark` - (Optional) The remarks of instance. (Default: "").
- `tag` - (Optional) A tag assigned to instance, which contains at most 63 characters and only support Chinese, English, numbers, '-', '_', and '.'. If it is not filled in or a empty string is filled in, then default tag will be assigned. (Default: Default).
- `security_group` - (Optional) The ID of the associated security group.
- `vpc_id` - (Optional, ForceNew) The ID of VPC linked to the instance. If not defined vpc_id, the instance will use the default VPC in the current region.
- `subnet_id` - (Optional, ForceNew) The ID of subnet. If defined vpc_id, the subnet_id is Required. If not defined vpc_id and subnet_id, the instance will use the default subnet in the current region.
- `private_ip` - (Optional, ForceNew) The private IP address assigned to the instance.
- `data_disks` - (Optional, ForceNew) Additional cloud data disks to attach to the instance. data_disks configurations only apply on resource creation. The count of data_disks can only be one.
- `network_interface` - (Optional, ForceNew) Additional network interface eips to attach to the instance. network_interface configurations only apply on resource creation. The count of network_interface can only be one. See network_interface below for details on attributes.
- `raid_type` -  (Optional, ForceNew) Types of RAID for local disk type instance. Possible values are raid1, raid0, raid10, raid5, no_raid.

### data_disks

The `data_disks` block supports:

- `size` - (Required, ForceNew) The size of the cloud data disk, range 20-8000, measured in GB (GigaByte).
- `type` - (Required, ForceNew) The type of the cloud data disk. Possible values are: cloud_normal for cloud disk, cloud_ssd for SSD cloud disk, cloud_rssd as RDMA-SSD cloud disk.

### network_interface

The `network_interface` block supports:

- `eip_bandwidth` - (Required, ForceNew) Maximum bandwidth to the elastic public network, measured in Mbps (Mega bit per second). The ranges for bandwidth are: 1-200 for pay by traffic, 1-800 for pay by bandwidth.
- `eip_internet_type` - (Required, ForceNew) Type of Elastic IP routes. Possible values are: international as international BGP IP and bgp as china mainland BGP IP.
- `eip_charge_mode` - (Required, ForceNew) Elastic IP charge mode.  Possible values are raid1, raid0, raid10, raid5 and no_raid.