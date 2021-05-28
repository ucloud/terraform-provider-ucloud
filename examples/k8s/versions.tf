terraform {
  required_version = ">= 0.12"
}

terraform {
  required_providers {
    ucloud = {
      source = "ucloud/ucloud"
      version = "~>1.27.0"
    }
  }
}