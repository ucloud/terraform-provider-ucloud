# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Query availability zone
data "ucloud_zones" "default" {
}

# Query image
data "ucloud_images" "default" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

# Create VPC
resource "ucloud_vpc" "default" {
  name = "tf-example-intranet-cluster"
  tag  = "tf-example"

  # vpc network
  cidr_blocks = ["192.168.0.0/16"]
}

# Create Subnet under the VPC
resource "ucloud_subnet" "default" {
  name = "tf-example-intranet-cluster"
  tag  = "tf-example"

  # subnet's network must be contained by vpc network
  # and a subnet must have least 8 ip addresses in it (netmask < 30).
  cidr_block = "192.168.1.0/24"

  vpc_id = ucloud_vpc.default.id
}

# Create a intranet cluster
resource "ucloud_instance" "intranet" {
  count = var.instance_count

  availability_zone = data.ucloud_zones.default.zones[0].id
  image_id          = data.ucloud_images.default.images[0].id
  instance_type     = "n-basic-2"
  root_password     = var.instance_password
  boot_disk_type    = "cloud_ssd"

  # we will put all the instances into same vpc and subnet,
  # so they can communicate with each other.
  vpc_id = ucloud_vpc.default.id

  subnet_id = ucloud_subnet.default.id

  name = "tf-example-intranet-cluster-${format(var.count_format, count.index + 1)}"
  tag  = "tf-example"
}

