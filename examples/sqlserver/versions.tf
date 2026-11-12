terraform {
  required_version = ">= 0.12"

  required_providers {
    ucloud = {
      source  = "ucloud/ucloud"
      version = "~>1.27.0"
    }
  }
}
