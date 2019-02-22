# Specify the provider and access details
provider "ucloud" {
  region = "${var.region}"
}

# Query availability zone
data "ucloud_zones" "default" {}


# Create database instance
resource "ucloud_db_instance" "master" {
  availability_zone  = "${data.ucloud_zones.default.zones.0.id}"
  name               = "tf-example-db-instance"
  instance_storage   = 20
  instance_type      = "mysql-ha-1"
  engine             = "mysql"
  engine_version     = "5.7"
  password           = "${var.db_password}"

  # Backup policy
  backup_begin_time = 4
  backup_count      = 6
  backup_date       = "0111110"
  backup_black_list = ["test.%"]
}