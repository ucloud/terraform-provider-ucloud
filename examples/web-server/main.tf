# Specify the provider and access details
provider "ucloud" {
  region = "${var.region}"
}

# Query availability zone
data "ucloud_zones" "default" {}

# Query image
data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

# Query default security group
data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

# Create a web server
resource "ucloud_instance" "web" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-2"
  root_password     = "${var.instance_password}"
  name              = "tf-example-web-server"
  tag               = "tf-example"

  # the default Web Security Group that UCloud recommend to users
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
}

# Create cloud disk
resource "ucloud_disk" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name              = "tf-example-web-server"
  disk_size         = 30
}

# Attach cloud disk to instance
resource "ucloud_disk_attachment" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  disk_id           = "${ucloud_disk.default.id}"
  instance_id       = "${ucloud_instance.web.id}"
}

# Create an eip
resource "ucloud_eip" "default" {
  bandwidth     = 2
  charge_mode   = "bandwidth"
  name          = "tf-example-web-server"
  tag           = "tf-example"
  internet_type = "bgp"
}

# Bind eip to instance
resource "ucloud_eip_association" "default" {
  resource_id = "${ucloud_instance.web.id}"
  eip_id      = "${ucloud_eip.default.id}"
}
