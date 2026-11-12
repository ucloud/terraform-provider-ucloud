provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "default" {
  name        = "tf-example-umongodb"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
  name       = "tf-example-umongodb"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.default.id
}

resource "ucloud_umongodb_replicaset" "example" {
  availability_zone = var.zone
  name              = "tf-example-umongodb"
  db_version        = var.db_version
  admin_password    = var.admin_password
  disk_space        = 20
  machine_type_id   = var.machine_type_id
  node_count        = 3

  vpc_id    = ucloud_vpc.default.id
  subnet_id = ucloud_subnet.default.id
}

data "ucloud_umongodb_instance" "example" {
  cluster_id = ucloud_umongodb_replicaset.example.id
}
