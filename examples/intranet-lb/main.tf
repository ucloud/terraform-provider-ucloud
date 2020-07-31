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

# Create intranet instances
resource "ucloud_instance" "intranet" {
  availability_zone = var.zone
  instance_type     = "n-basic-2"
  image_id          = data.ucloud_images.default.images[0].id
  root_password     = var.instance_password

  name  = "tf-example-intranet-lb-${format(var.count_format, count.index + 1)}"
  tag   = "tf-example"
  count = var.instance_count
}

# Create Load Balancer
resource "ucloud_lb" "default" {
  name     = "tf-example-intranet-lb"
  tag      = "tf-example"
  internal = true
}

# Create Load Balancer Listener with tcp protocol
resource "ucloud_lb_listener" "default" {
  name             = "tf-example-intranet-lb"
  listen_type      = "packets_transmit"
  load_balancer_id = ucloud_lb.default.id
  protocol         = "tcp"
}

# Attach instances to Load Balancer
resource "ucloud_lb_attachment" "default" {
  load_balancer_id = ucloud_lb.default.id
  listener_id      = ucloud_lb_listener.default.id
  resource_id      = ucloud_instance.intranet[count.index].id
  port             = 1024
  count            = var.instance_count
}
