output "instance_name_list" {
  value = ucloud_instance.intranet.*.name
}

output "instance_id_list" {
  value = ucloud_instance.intranet.*.id
}

output "private_ip_list" {
  value = ucloud_instance.intranet.*.private_ip
}

