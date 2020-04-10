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

# Create Security Group
resource "ucloud_security_group" "default" {
  name = "tf-example-two-tier"
  tag  = "tf-example"

  # HTTP access from LAN
  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "0.0.0.0/0"
    policy     = "accept"
  }

  # HTTPS access from LAN
  rules {
    port_range = "443"
    protocol   = "tcp"
    cidr_block = "0.0.0.0/0"
    policy     = "accept"
  }
}

# Create VPC
resource "ucloud_vpc" "default" {
  name = "tf-example-two-tier"
  tag  = "tf-example"

  # vpc network
  cidr_blocks = ["192.168.0.0/16"]
}

# Create Subnet under the VPC
resource "ucloud_subnet" "default" {
  name = "tf-example-two-tier"
  tag  = "tf-example"

  # subnet's network must be contained by vpc network
  # and a subnet must have least 8 ip addresses in it (netmask < 30).
  cidr_block = "192.168.1.0/24"

  vpc_id = ucloud_vpc.default.id
}

# Create web servers
resource "ucloud_instance" "web" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  instance_type     = "n-basic-2"
  image_id          = data.ucloud_images.default.images[0].id
  root_password     = var.instance_password
  boot_disk_type    = "cloud_ssd"

  name = "tf-example-two-tier-${format(var.count_format, count.index + 1)}"
  tag  = "tf-example"

  # we will put all the instances into same vpc and subnet,
  # so they can communicate with each other.
  vpc_id = ucloud_vpc.default.id

  subnet_id = ucloud_subnet.default.id

  # this security group allows HTTP and HTTPS access
  security_group = ucloud_security_group.default.id

  count = var.instance_count
}

# Create Load Balancer
resource "ucloud_lb" "default" {
  name = "tf-example-two-tier"
  tag  = "tf-example"

  # we will put all the instances into same vpc and subnet,
  # so they can communicate with each other.
  vpc_id = ucloud_vpc.default.id

  subnet_id = ucloud_subnet.default.id
}

# Create Load Balancer Listener with http protocol
resource "ucloud_lb_listener" "default" {
  load_balancer_id = ucloud_lb.default.id
  protocol         = "http"
  port             = 80
}

# Attach instances to Load Balancer
resource "ucloud_lb_attachment" "default" {
  load_balancer_id = ucloud_lb.default.id
  listener_id      = ucloud_lb_listener.default.id
  resource_id      = ucloud_instance.web[count.index].id
  port             = 80
  count            = var.instance_count
}

# Create Load Balancer Listener Rule
resource "ucloud_lb_rule" "default" {
  load_balancer_id = ucloud_lb.default.id
  listener_id      = ucloud_lb_listener.default.id
  backend_ids      = ucloud_lb_attachment.default.*.id
  domain           = "www.ucloud.cn"
}

# Create an eip
resource "ucloud_eip" "default" {
  bandwidth     = 2
  charge_mode   = "bandwidth"
  name          = "tf-example-two-tier"
  tag           = "tf-example"
  internet_type = "bgp"
}

# Bind eip to Load Balancer
resource "ucloud_eip_association" "default" {
  resource_id = ucloud_lb.default.id
  eip_id      = ucloud_eip.default.id
}

