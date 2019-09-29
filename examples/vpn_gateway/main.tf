provider "ucloud" {
  region = var.region
}

resource "ucloud_vpc" "foo" {
  name        = "tf-example-vpn-connection"
  tag         = "tf-example"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-example-vpn-connection"
  tag        = "tf-example"
  cidr_block = "192.168.1.0/24"
  vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_eip" "foo" {
  name          = "tf-example-vpn-connection"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-example"
}

resource "ucloud_vpn_gateway" "foo" {
  vpc_id = ucloud_vpc.foo.id
  grade  = "standard"
  eip_id = ucloud_eip.foo.id
  name   = "tf-example-vpn-connection"
  tag    = "tf-example"
}

resource "ucloud_vpn_customer_gateway" "foo" {
  ip_address = "10.0.0.1"
  name       = "tf-example-vpn-connection"
  tag        = "tf-example"
}

resource "ucloud_vpn_connection" "foo" {
  vpn_gateway_id      = ucloud_vpn_gateway.foo.id
  customer_gateway_id = ucloud_vpn_customer_gateway.foo.id
  vpc_id              = ucloud_vpc.foo.id
  name                = "tf-example-vpn-connection"
  tag                 = "tf-example"
  remark              = "test"
  ike_config {
    pre_shared_key           = "password_2019"
    exchange_mode            = "aggressive"
    encryption_algorithm     = "aes192"
    authentication_algorithm = "md5"
    dh_group                 = 14
    sa_life_time             = 10000
  }

  ipsec_config {
    local_subnet_ids         = [ucloud_subnet.foo.id]
    remote_subnets           = ["10.0.0.0/24"]
    protocol                 = "ah"
    encryption_algorithm     = "aes192"
    authentication_algorithm = "md5"
    sa_life_time             = 10000
  }
}