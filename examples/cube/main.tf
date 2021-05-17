# Specify the provider and access details
provider "ucloud" {
  region = var.region
}

data "ucloud_vpcs" "default" {
  name_regex = "DefaultVPC"
}
data "ucloud_subnets" "default" {
  vpc_id = data.ucloud_vpcs.default.vpcs.0.id
}

resource "ucloud_cube_pod" "foo" {
  availability_zone = var.zone
  name              = "tf-acc-cube-pod-basic"
  tag               = "tf-acc"
  vpc_id            = data.ucloud_vpcs.default.vpcs.0.id
  subnet_id         = data.ucloud_subnets.default.subnets.0.id
  pod               = file("cube_pod.yml")
}

# Create an eip
resource "ucloud_eip" "default" {
  bandwidth     = 2
  charge_mode   = "bandwidth"
  name          = "tf-example-web-server"
  tag           = "tf-example"
  internet_type = "bgp"
}

# Bind eip to instance
resource "ucloud_eip_association" "default" {
  resource_id = ucloud_cube_pod.foo.id
  eip_id      = ucloud_eip.default.id
}

