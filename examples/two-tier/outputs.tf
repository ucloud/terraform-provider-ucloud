output "instance_name_list" {
  value = "${ucloud_instance.web.*.name}"
}

output "instance_id_list" {
  value = "${ucloud_instance.web.*.id}"
}

output "load_balancer_id" {
  value = "${ucloud_lb.default.id}"
}

output "public_ip" {
  value = "${ucloud_eip.default.public_ip}"
}
