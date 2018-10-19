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
  name = "tf-example-disk"
  tag  = "tf-example"

  # allow all access from WAN
  rules {
    port_range = "1-65535"
    protocol   = "TCP"
    cidr_block = "0.0.0.0/0"
    policy     = "ACCEPT"
  }
}

resource "ucloud_disk" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name              = "tf-example-disk"
  disk_size         = 10
}

resource "ucloud_instance" "web" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  instance_type     = "n-standard-1"

  image_id      = "${data.ucloud_images.default.images.0.id}"
  root_password = "${var.instance_password}"

  # this ecurity group to allow all access from WAN
  security_group = "${ucloud_security_group.default.id}"

  name = "tf-example-disk"
  tag  = "tf-example"
}

resource "ucloud_disk_attachment" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  disk_id           = "${ucloud_disk.default.id}"
  instance_id       = "${ucloud_instance.web.id}"
}
