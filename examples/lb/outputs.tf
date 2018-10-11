output "instance_id" {
  value = "${ucloud_instance.web.id}"
}

output "load_balancer_id" {
  value = "${ucloud_lb.web.id}"
}
