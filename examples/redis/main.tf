# Specify the provider and access details
provider "ucloud" {
  region = "${var.region}"
}

# Query availability zone
data "ucloud_zones" "default" {}

# Create VPC
resource "ucloud_vpc" "default" {
  name = "tf-example-redis"
  tag  = "tf-example"

  # vpc network
  cidr_blocks = ["192.168.0.0/16"]
}

# Create Subnet under the VPC
resource "ucloud_subnet" "default" {
  name = "tf-example-redis"
  tag  = "tf-example"

  # subnet's network must be contained by vpc network
  # and a subnet must have least 8 ip addresses in it (netmask < 30).
  cidr_block = "192.168.1.0/24"

  vpc_id = "${ucloud_vpc.default.id}"
}

# Create redis instance
resource "ucloud_redis_instance" "master" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  engine_version    = "4.0"
  instance_type     = "redis-master-2"
  password          = "${var.redis_password}"
  name              = "tf-example-redis"
  tag               = "tf-example"

  vpc_id    = "${ucloud_vpc.default.id}"
  subnet_id = "${ucloud_subnet.default.id}"
}
