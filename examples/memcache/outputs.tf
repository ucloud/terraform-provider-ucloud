output "memcache_instance_id" {
  value = "${ucloud_memcache_instance.master.id}"
}

output "private_ip" {
  value = "${ucloud_memcache_instance.master.ip_set.0.ip}"
}

output "port" {
  value = "${ucloud_memcache_instance.master.ip_set.0.port}"
}
