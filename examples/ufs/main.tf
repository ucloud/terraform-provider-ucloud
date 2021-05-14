# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_ufs" "foo" {
  name          = "tf-acc-ufs-basic"
  remark        = "test"
  tag           = "tf-acc"
  size          = 600
  storage_type  = "Basic"
  protocol_type = "NFSv4"
}