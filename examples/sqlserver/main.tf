# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

# Create a SQL Server instance on the nvme machine type
resource "ucloud_sqlserver_instance" "master" {
  availability_zone = var.zone
  name              = "tf-example-sqlserver"
  db_version        = "sqlserver-2019"
  cpu               = 2
  memory            = 4000   # MB
  disk_space        = 50     # GB
  password          = var.db_password
}
