package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	unetapi "github.com/ucloud/ucloud-sdk-go/services/unet"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudNatGW_basic(t *testing.T) {
	var val vpcapi.NatGatewayDataSet
	var vpcSet vpcapi.VPCInfo
	var subnetSet vpcapi.SubnetInfo
	var eipSet unetapi.UnetEIPSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_nat_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNatGWDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccNatGWConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGWExists("ucloud_nat_gateway.foo", &val),
					testAccCheckVPCExists("ucloud_vpc.foo", &vpcSet),
					testAccCheckSubnetExists("ucloud_subnet.foo", &subnetSet),
					testAccCheckSubnetExists("ucloud_subnet.bar", &subnetSet),
					testAccCheckEIPExists("ucloud_eip.foo", &eipSet),
					testAccCheckNatGWAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "name", "tf-acc-nat-gateway-basic"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "enable_white_list", "false"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "white_list.#", "2"),
				),
			},
			{
				Config: testAccNatGWConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGWExists("ucloud_nat_gateway.foo", &val),
					testAccCheckVPCExists("ucloud_vpc.foo", &vpcSet),
					testAccCheckSubnetExists("ucloud_subnet.foo", &subnetSet),
					testAccCheckEIPExists("ucloud_eip.foo", &eipSet),
					testAccCheckNatGWAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "name", "tf-acc-nat-gateway-basic"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "enable_white_list", "true"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("ucloud_nat_gateway.foo", "white_list.#", "1"),
				),
			},
		},
	})
}

const testAccNatGWConfig = `
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-nat-gateway-basic"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-nat-gateway-basic"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_subnet" "bar" {
  name       = "tf-acc-nat-gateway-basic"
  tag        = "tf-acc"
  cidr_block = "192.168.2.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-nat-gateway-basic"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
  type = "recommend_web"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-basic"
  tag               = "tf-acc"
}

resource "ucloud_instance" "bar" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.bar.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id, ucloud_subnet.bar.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateway-basic"
  tag               = "tf-acc"
  white_list        = [ucloud_instance.foo.id, ucloud_instance.bar.id]
  enable_white_list = false
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}
`

const testAccNatGWConfigUpdate = `
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-nat-gateway-basic"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-nat-gateway-basic"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

resource "ucloud_eip" "foo" {
  name          = "tf-acc-nat-gateway-basic"
  bandwidth     = 1
  internet_type = "bgp"
  charge_mode   = "bandwidth"
  tag           = "tf-acc"
}

data "ucloud_security_groups" "foo" {
  type = "recommend_web"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_id         = ucloud_subnet.foo.id
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  charge_type       = "dynamic"
  name              = "tf-acc-nat-gateway-basic"
  tag               = "tf-acc"
}

resource "ucloud_nat_gateway" "foo" {
  vpc_id            = ucloud_vpc.foo.id
  subnet_ids        = [ucloud_subnet.foo.id]
  eip_id            = ucloud_eip.foo.id
  name              = "tf-acc-nat-gateway-basic"
  tag               = "tf-acc"
  enable_white_list = true
  white_list        = [ucloud_instance.foo.id]
  security_group    = data.ucloud_security_groups.foo.security_groups.0.id
}
`
