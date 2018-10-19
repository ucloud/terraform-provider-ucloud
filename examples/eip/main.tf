# Specify the provider and access details
provider "ucloud" {
  region = "${var.region}"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "Base"
}

resource "ucloud_security_group" "default" {
  name = "tf-example-eip"
  tag  = "tf-example"

  rules {
    port_range = "80"
    protocol   = "TCP"
    cidr_block = "192.168.0.0/16"
    policy     = "ACCEPT"
  }
}

resource "ucloud_eip" "default" {
  bandwidth            = 2
  internet_charge_mode = "Bandwidth"
  name                 = "tf-example-eip"
  tag                  = "tf-example"
}

resource "ucloud_instance" "web" {
  instance_type     = "n-standard-1"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"

  data_disk_size = 50
  root_password  = "${var.instance_password}"
  security_group = "${ucloud_security_group.default.id}"

  name = "tf-example-eip"
  tag  = "tf-example"
}

resource "ucloud_eip_association" "default" {
  resource_type = "instance"
  resource_id   = "${ucloud_instance.web.id}"
  eip_id        = "${ucloud_eip.default.id}"
}
