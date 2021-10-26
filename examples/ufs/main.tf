# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

data "ucloud_vpcs" "default" {
}

data "ucloud_subnets" "default" {
  vpc_id = data.ucloud_vpcs.default.vpcs[0].id
}

resource "ucloud_ufs_volume" "foo" {
  name          = "tf-acc-ufs-basic"
  remark        = "test"
  tag           = "tf-acc"
  size          = 600
  storage_type  = "Basic"
  protocol_type = "NFSv4"
}

resource "ucloud_ufs_volume_mount_point" "foo" {
  name      = "tf-acc-ufs-mount-point-basic"
  volume_id = ucloud_ufs_volume.foo.id
  vpc_id    = data.ucloud_vpcs.default.vpcs[0].id
  subnet_id = data.ucloud_subnets.default.subnets[0].id
}