terraform {
  required_providers {
    ucloud = {
      source  = "ucloud/ucloud"
      version = "~>1.38.3"
    }
  }
}

provider "ucloud" {
  region = var.region
}

# Create a ssd cloud disk. The snapshot service has to be enabled here, it cannot be
# enabled on an existing disk, and without it no snapshot can be created from this disk.
resource "ucloud_disk" "default" {
  availability_zone = var.zone
  name              = "tf-example-disk-snapshot"
  tag               = "tf-example"
  disk_size         = 20
  disk_type         = "ssd_data_disk"
  snapshot_service  = true
}

# Take a snapshot of the cloud disk
resource "ucloud_disk_snapshot" "default" {
  availability_zone = var.zone
  disk_id           = ucloud_disk.default.id
  name              = "tf-example-disk-snapshot"
}

# Clone a new cloud disk from the snapshot. The disk_type is left unset on purpose,
# the cloned disk inherits the type of its source snapshot, which is ssd_data_disk here.
resource "ucloud_disk" "cloned" {
  availability_zone = var.zone
  name              = "tf-example-disk-snapshot-cloned"
  tag               = "tf-example"
  disk_size         = 20
  snapshot_id       = ucloud_disk_snapshot.default.id
}

# Query the snapshots of the source disk
data "ucloud_disk_snapshots" "default" {
  availability_zone = var.zone
  disk_id           = ucloud_disk.default.id

  depends_on = [ucloud_disk_snapshot.default]
}

output "snapshot_id" {
  value = ucloud_disk_snapshot.default.id
}

output "cloned_disk_type" {
  value = ucloud_disk.cloned.disk_type
}

output "snapshot_total_count" {
  value = data.ucloud_disk_snapshots.default.total_count
}
