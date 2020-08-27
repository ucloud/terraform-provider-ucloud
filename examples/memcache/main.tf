# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Create VPC
resource "ucloud_vpc" "default" {
  name = "tf-example-memcache"
  tag  = "tf-example"

  # vpc network
  cidr_blocks = ["192.168.0.0/16"]
}

# Create Subnet under the VPC
resource "ucloud_subnet" "default" {
  name = "tf-example-memcache"
  tag  = "tf-example"

  # subnet's network must be contained by vpc network
  # and a subnet must have least 8 ip addresses in it (netmask < 30).
  cidr_block = "192.168.1.0/24"

  vpc_id = ucloud_vpc.default.id
}

# Create memcache instance
resource "ucloud_memcache_instance" "master" {
  availability_zone = var.zone
  name              = "tf-example-memcache"
  instance_type     = "memcache-master-2"

  vpc_id    = ucloud_vpc.default.id
  subnet_id = ucloud_subnet.default.id
}

