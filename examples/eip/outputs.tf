output "instance_id" {
  value = "${ucloud_instance.web.id}"
}

output "elastic_ip" {
  value = "${ucloud_eip.default.ip_set.0.ip}"
}
