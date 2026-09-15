# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-example-subnet"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_network_interface" "foo" {
  name      = "tf-example-network-interface"
  tag       = "tf-example"
  remark    = "tf-example-network-interface"
  vpc_id    = ucloud_vpc.foo.id
  subnet_id = ucloud_subnet.foo.id
}
