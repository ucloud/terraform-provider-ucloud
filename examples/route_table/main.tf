# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpc"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_route_table" "foo" {
  name   = "tf-example-route-table"
  tag    = "tf-example"
  remark = "tf-example-route-table"
  vpc_id = ucloud_vpc.foo.id
}
