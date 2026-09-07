terraform {
  required_providers {
    ucloud = {
      source  = "ucloud/ucloud"
      version = "1.39.6"
    }
  }
}

variable "bucket_name" {
  type = string
}

variable "bucket_type" {
  type = string
}

provider "ucloud" {}

resource "ucloud_us3_bucket" "compat" {
  name = var.bucket_name
  type = var.bucket_type
}
