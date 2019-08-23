output "instance_name_list" {
  value = ucloud_instance.intranet.*.name
}

output "instance_id_list" {
  value = ucloud_instance.intranet.*.id
}

output "load_balancer_id" {
  value = ucloud_lb.default.id
}

output "lb_private_ip" {
  value = ucloud_lb.default.private_ip
}


