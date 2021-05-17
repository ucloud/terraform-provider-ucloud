# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_us3_bucket" "foo" {
  name = "tf-acc-us3-bucket-basic"
  type = "public"
  tag  = "tf-acc"
}