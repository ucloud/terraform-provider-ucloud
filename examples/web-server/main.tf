# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Query image
data "ucloud_images" "default" {
  availability_zone = var.zone
  name_regex        = "^CentOS 8.[1-2] 64"
  image_type        = "base"
}

# Query default security group
data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

# Create a web server
resource "ucloud_instance" "web" {
  availability_zone = var.zone
  image_id          = data.ucloud_images.default.images[0].id
  instance_type     = "n-basic-2"
  root_password     = var.instance_password
  name              = "tf-example-web-server"
  tag               = "tf-example"
  boot_disk_type    = "cloud_ssd"

  # the default Web Security Group that UCloud recommend to users
  security_group = data.ucloud_security_groups.default.security_groups[0].id

  # create cloud data disk attached to instance
  data_disks {
    size = 20
    type = "cloud_ssd"
  }
  delete_disks_with_instance = true
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
  resource_id = ucloud_instance.web.id
  eip_id      = ucloud_eip.default.id
}

