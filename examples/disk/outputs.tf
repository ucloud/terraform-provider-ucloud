output "instance_id" {
  value = "${ucloud_instance.web.id}"
}

output "disk_id" {
  value = "${ucloud_disk.default.id}"
}
