provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-example-nat-gateway"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-example-nat-gateway"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_subnet" "bar" {
  name       = "tf-example-nat-gateway"
  tag        = "tf-example"
  cidr_block = "192.168.2.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_eip" "foo" {
  name          = "tf-example-nat-gateway"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-example"
}

data "ucloud_security_groups" "foo" {
  type = "recommend_web"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = data.ucloud_zones.default.zones.0.id
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = data.ucloud_zones.default.zones.0.id
  image_id          = data.ucloud_images.default.images.0.id
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-example-nat-gateway"
  tag               = "tf-example"
  count             = 2
}

resource "ucloud_instance" "bar" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.bar.id
  availability_zone = data.ucloud_zones.default.zones.0.id
  image_id          = data.ucloud_images.default.images.0.id
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-example-nat-gateway"
  tag               = "tf-example"
  count             = 2
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id, ucloud_subnet.bar.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-example-nat-gateway"
  tag               = "tf-example"
  white_list        = [ucloud_instance.foo.0.id, ucloud_instance.foo.1.id, ucloud_instance.bar.0.id, ucloud_instance.bar.1.id]
  enable_white_list = true
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}

resource "ucloud_nat_gateway_rule" "foo" {
  nat_gateway_id = ucloud_nat_gateway.foo.id
  protocol       = "tcp"
  src_eip_id     = ucloud_eip.foo.id
  src_port_range = "80"
  dst_ip         = ucloud_instance.foo.0.private_ip
  dst_port_range = "88"
  name           = "tf-acc-nat-gateway-rule-update"
}

resource "ucloud_nat_gateway_rule" "bar" {
  nat_gateway_id = ucloud_nat_gateway.foo.id
  protocol       = "tcp"
  src_eip_id     = ucloud_eip.foo.id
  src_port_range = "90-100"
  dst_ip         = ucloud_instance.foo.1.private_ip
  dst_port_range = "90-100"
  name           = "tf-acc-nat-gateway-rule-update"
}