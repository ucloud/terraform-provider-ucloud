# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Create a MySQL instance on the nvme machine type
resource "ucloud_udb_mysql_instance" "master" {
  availability_zone = var.zone
  name              = "tf-example-mysql"
  db_version        = "mysql-8.0"
  machine_type      = "o.mysql2m.small"
  disk_space        = 20
  password          = var.db_password

  # HA is the default; set "Normal" for a standalone instance
  instance_mode = "HA"
}
