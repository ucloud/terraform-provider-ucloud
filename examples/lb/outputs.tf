output "instance_id_list" {
  value = "${ucloud_instance.web.*.id}"
}

output "load_balancer_id" {
  value = "${ucloud_lb.default.id}"
}
