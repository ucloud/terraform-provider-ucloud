output "instance_name_list" {
  value = "${ucloud_instance.web.*.name}"
}

output "instance_id_list" {
  value = "${ucloud_instance.web.*.id}"
}
