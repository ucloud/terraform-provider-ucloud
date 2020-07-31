# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Query image
data "ucloud_images" "default" {
  availability_zone = var.zone
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

# Create Security Group
resource "ucloud_security_group" "default" {
  name = "tf-example-lb"
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

# Create web servers
resource "ucloud_instance" "web" {
  availability_zone = var.zone
  instance_type     = "n-basic-2"
  boot_disk_type    = "cloud_ssd"

  image_id      = data.ucloud_images.default.images[0].id
  root_password = var.instance_password

  # this security group allows HTTP and HTTPS access
  security_group = ucloud_security_group.default.id

  name  = "tf-example-lb-${format(var.count_format, count.index + 1)}"
  tag   = "tf-example"
  count = var.instance_count
}

# Create Load Balancer
resource "ucloud_lb" "default" {
  name = "tf-example-lb"
  tag  = "tf-example"
}

# Create Load Balancer Listener with https protocol
resource "ucloud_lb_listener" "default" {
  name             = "tf-example-lb"
  load_balancer_id = ucloud_lb.default.id
  protocol         = "https"
  port             = 443
}

# Create SSL certificate
resource "ucloud_lb_ssl" "default" {
  name        = "tf-example-lb-ssl-attachment"
  private_key = file("private.key")
  user_cert   = file("user.crt")
  ca_cert     = file("ca.crt")
}

# Attach SSL certificate to Load Balancer Listener
resource "ucloud_lb_ssl_attachment" "default" {
  load_balancer_id = ucloud_lb.default.id
  listener_id      = ucloud_lb_listener.default.id
  ssl_id           = ucloud_lb_ssl.default.id
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
  name          = "tf-example-lb"
  tag           = "tf-example"
  internet_type = "bgp"
}

# Bind eip to Load Balancer
resource "ucloud_eip_association" "default" {
  resource_id = ucloud_lb.default.id
  eip_id      = ucloud_eip.default.id
}

