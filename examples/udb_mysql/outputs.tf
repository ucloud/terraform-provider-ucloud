output "mysql_instance_id" {
  value = ucloud_udb_mysql_instance.master.id
}

output "private_ip" {
  value = ucloud_udb_mysql_instance.master.private_ip
}
