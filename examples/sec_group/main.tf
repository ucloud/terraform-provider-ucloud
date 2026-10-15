provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-example-sec-group"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_sec_group" "foo" {
  name   = "tf-example-sec-group"
  vpc_id = ucloud_vpc.foo.id
  remark = "managed by terraform"
}

resource "ucloud_sec_group_rule" "ssh" {
  sec_group_id  = ucloud_sec_group.foo.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = "22"
  ip_range      = "10.0.0.0/8"
  rule_action   = "Accept"
  priority      = 50
  remark        = "ssh"
}

resource "ucloud_sec_group_rule" "web" {
  sec_group_id  = ucloud_sec_group.foo.id
  direction     = "Ingress"
  protocol_type = "TCP"
  dst_port      = "80,443"
  ip_range      = "0.0.0.0/0"
  rule_action   = "Accept"
  priority      = 60
  remark        = "web"
}

# ICMP carries no port, so dst_port is left out.
resource "ucloud_sec_group_rule" "ping" {
  sec_group_id  = ucloud_sec_group.foo.id
  direction     = "Egress"
  protocol_type = "ICMP"
  ip_range      = "0.0.0.0/0"
  rule_action   = "Accept"
  priority      = 100
}
