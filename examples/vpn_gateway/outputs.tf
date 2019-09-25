output "vpn_gateway_id" {
  value = ucloud_vpn_gateway.foo.id
}

output "vpn_customer_gateway_id" {
  value = ucloud_vpn_customer_gateway.foo.id
}

output "ucloud_vpn_connection_id" {
  value = ucloud_vpn_connection.foo.id
}

output "public_ip" {
  value = ucloud_eip.foo.public_ip
}


