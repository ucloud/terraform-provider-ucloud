output "sqlserver_instance_id" {
  value = ucloud_sqlserver_instance.master.id
}

output "private_ip" {
  value = ucloud_sqlserver_instance.master.private_ip
}
