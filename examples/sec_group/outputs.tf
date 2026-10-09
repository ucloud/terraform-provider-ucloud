output "sec_group_id" {
  value = ucloud_sec_group.foo.id
}

output "rule_id_list" {
  value = [
    ucloud_sec_group_rule.ssh.id,
    ucloud_sec_group_rule.web.id,
    ucloud_sec_group_rule.ping.id,
  ]
}
