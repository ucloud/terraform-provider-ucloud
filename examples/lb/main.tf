# Specify the provider and access details
provider "ucloud" {
  region = "${var.region}"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_security_group" "default" {
  name = "tf-example-lb"
  tag  = "tf-example"

  # HTTP access from LAN
  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }

  # HTTPS access from LAN
  rules {
    port_range = "443"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }
}

resource "ucloud_lb" "default" {
  name = "tf-example-lb"
  tag  = "tf-example"
}

resource "ucloud_lb_listener" "default" {
  load_balancer_id = "${ucloud_lb.default.id}"
  protocol         = "https"
}

resource "ucloud_lb_ssl" "default" {
  name        = "tf-example-lb-ssl-attachment"
  private_key = "${file("private.key")}"
  user_cert   = "${file("user.crt")}"
  ca_cert     = "${file("ca.crt")}"
}

resource "ucloud_lb_ssl_attachment" "default" {
  load_balancer_id = "${ucloud_lb.default.id}"
  listener_id      = "${ucloud_lb_listener.default.id}"
  ssl_id           = "${ucloud_lb_ssl.default.id}"
}

resource "ucloud_instance" "web" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  instance_type     = "n-standard-1"

  image_id      = "${data.ucloud_images.default.images.0.id}"
  root_password = "${var.instance_password}"

  # this security group allows HTTP and HTTPS access
  security_group = "${ucloud_security_group.default.id}"

  name  = "tf-example-lb-${format(var.count_format, count.index+1)}"
  tag   = "tf-example"
  count = "${var.count}"
}

resource "ucloud_lb_attachment" "default" {
  load_balancer_id = "${ucloud_lb.default.id}"
  listener_id      = "${ucloud_lb_listener.default.id}"
  resource_id      = "${ucloud_instance.web.*.id[count.index]}"
  port             = 80
  count            = "${var.count}"
}

resource "ucloud_lb_rule" "default" {
  load_balancer_id = "${ucloud_lb.default.id}"
  listener_id      = "${ucloud_lb_listener.default.id}"
  backend_ids      = ["${ucloud_lb_attachment.default.*.id}"]
  domain           = "www.ucloud.cn"
}
