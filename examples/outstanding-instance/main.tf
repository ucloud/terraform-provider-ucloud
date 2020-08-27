# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Query image
data "ucloud_images" "default" {
  availability_zone = var.zone
  name_regex        = "^高内核CentOS 7.6 64位"
  image_type        = "base"
}

# Query default security group
data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

# Create a outstanding instance
resource "ucloud_instance" "outstanding" {
  availability_zone = var.zone
  image_id          = data.ucloud_images.default.images[0].id
  instance_type     = "o-highcpu-2"
  root_password     = var.instance_password
  name              = "tf-example-outstanding-instance"
  tag               = "tf-example"
  boot_disk_type    = "cloud_rssd"
  min_cpu_platform  = "Amd/Epyc2"

  # the default Web Security Group that UCloud recommend to users
  security_group = data.ucloud_security_groups.default.security_groups[0].id

  # create cloud data disk attached to instance
  data_disks {
    size = 20
    type = "cloud_rssd"
  }
  delete_disks_with_instance = true
}

