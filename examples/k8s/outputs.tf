output "api_server" {
  value = ucloud_uk8s_cluster.foo.api_server
}

output "kubeconfig" {
  value     = ucloud_uk8s_cluster.foo.kubeconfig
  sensitive = true
}

output "external_kubeconfig" {
  value     = ucloud_uk8s_cluster.foo.external_kubeconfig
  sensitive = true
}
