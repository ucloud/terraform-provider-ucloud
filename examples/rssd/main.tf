terraform {
  required_providers {
    ucloud = {
      source = "ucloud/ucloud"
    }
  }
}

provider "ucloud" {
  region = var.region
}

# Query image
data "ucloud_images" "default" {
  availability_zone = var.zone
  name_regex        = "^CentOS 8.[1-2] 64"
  image_type        = "base"
}

# Create a uhost
resource "ucloud_instance" "default" {
  availability_zone = var.zone
  image_id          = data.ucloud_images.default.images[0].id
  instance_type     = "o-standard-2"
  root_password     = var.instance_password
  name              = "tf-example-rssd"
  tag               = "tf-example"
  boot_disk_type    = "cloud_rssd"

  delete_disks_with_instance = true
}

# Create a rssd udisk
resource "ucloud_disk" "default" {
  availability_zone = var.zone
  name              = "tf-example-rssd"
  disk_size         = 10
  disk_type         = "rssd_data_disk"
  rdma_cluster_id   = ucloud_instance.default.rdma_cluster_id
}

# attach cloud disk to instance
resource "ucloud_disk_attachment" "default" {
  availability_zone = var.zone
  disk_id           = ucloud_disk.default.id
  instance_id       = ucloud_instance.default.id
}
