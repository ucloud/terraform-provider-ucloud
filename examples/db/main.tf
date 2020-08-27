# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Create database instance
resource "ucloud_db_instance" "master" {
  availability_zone = var.zone
  name              = "tf-example-db"
  instance_storage  = 20
  instance_type     = "mysql-ha-nvme-2"
  engine            = "mysql"
  engine_version    = "5.7"
  password          = var.db_password

  # Backup policy
  backup_begin_time = 4
  backup_count      = 6
  backup_date       = "0111110"
  backup_black_list = ["test.%"]
}

