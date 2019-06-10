provider "ucloud" {
  region     = "cn-bj2"
  profile    = "default"
  project_id = "org-x3voeo"
}

# Query default security group
data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

# Query image
data "ucloud_images" "default" {
  availability_zone = "cn-bj2-02"
  name_regex        = "^Windows 2016 64"
  image_type        = "base"
}

#   availability_zone = "${data.ucloud_zones.default.zones.0.id}"
#   name_regex        = "^CentOS 7.[1-2] 64"
#   image_type        = "base"
# Create web instance
resource "ucloud_instance" "web" {
  name              = "tf-example-instance"
  tag               = "tf-example"
  availability_zone = "cn-bj2-05"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  root_password     = "${var.instance_password}"
  boot_disk_type    = "cloud_ssd"

  #   # use local disk as data disk
  #   data_disk_size = "70"
  #   data_disk_type = "local_normal"

  # the default Web Security Group that UCloud recommend to users
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"

  #   count = "${var.count}"
}

# variable "count" {
#   default = 3
# }

variable "instance_password" {
  default = "Wa123456"
}
