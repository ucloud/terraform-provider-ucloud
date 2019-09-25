output "nat_gateway_id" {
  value = ucloud_nat_gateway.foo.id
}

output "instance_id_list" {
  value = ucloud_instance.foo.*.id
}

output "public_ip" {
  value = ucloud_eip.foo.public_ip
}


