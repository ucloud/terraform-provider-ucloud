output "cluster_id" {
  value = ucloud_umongodb_replicaset.example.id
}

output "connect_url" {
  value = data.ucloud_umongodb_instance.example.connect_url
}
