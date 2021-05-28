# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-acc-uk8s-cluster"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}
resource "ucloud_subnet" "foo" {
  name       = "tf-acc-uk8s-cluster"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
  vpc_id               = ucloud_vpc.foo.id
  subnet_id            = ucloud_subnet.foo.id
  name                 = "tf-acc-uk8s-cluster-basic"
  service_cidr         = "172.16.0.0/16"
  password             = var.password
  charge_type          = "dynamic"
  master_instance_type = "n-basic-2"
  master {
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
  master {
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
  master {
    availability_zone = data.ucloud_zones.default.zones.0.id
  }

  nodes {
    instance_type     = "n-basic-2"
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
  nodes {
    instance_type     = "n-basic-8"
    availability_zone = data.ucloud_zones.default.zones.0.id
  }
}
