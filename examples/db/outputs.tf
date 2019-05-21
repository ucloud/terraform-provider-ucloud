output "db_instance_id" {
  value = "${ucloud_db_instance.master.id}"
}

output "private_ip" {
  value = "${ucloud_db_instance.master.private_ip}"
}
