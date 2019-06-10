data "ucloud_images" "default" {
  availability_zone = "cn-bj2-04"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

# Create Load Balancer
resource "ucloud_lb" "web" {
  name = "tf-example-lb"
  tag  = "tf-example"
}

# Create Load Balancer Listener with http protocol
resource "ucloud_lb_listener" "default" {
  load_balancer_id = ucloud_lb.web.id
  protocol         = "http"
}

# Create web server
resource "ucloud_instance" "web" {
  instance_type     = "n-basic-2"
  availability_zone = "cn-bj2-04"

  root_password     = "wA1234567"
  image_id          = data.ucloud_images.default.images[0].id

name              = "tf-example-lb"
tag               = "tf-example"
}

# Attach instances to Load Balancer
resource "ucloud_lb_attachment" "example" {
load_balancer_id = ucloud_lb.web.id
listener_id      = ucloud_lb_listener.default.id
resource_id      = ucloud_instance.web.id
port             = 80
}
