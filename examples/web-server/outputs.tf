output "instance_name" {
  value = "${ucloud_instance.web.name}"
}

output "instance_id" {
  value = "${ucloud_instance.web.id}"
}

output "public_ip" {
  value = "${ucloud_eip.default.public_ip}"
}
