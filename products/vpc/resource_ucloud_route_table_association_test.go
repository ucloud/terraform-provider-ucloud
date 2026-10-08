package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudRouteTableAssociation_basic(t *testing.T) {
	var subnet vpcapi.SubnetInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_route_table_association.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableAssociationDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists("ucloud_route_table_association.foo", &subnet),
					testAccCheckRouteTableAssociationAttributes(&subnet),
					resource.TestCheckResourceAttrPair("ucloud_route_table_association.foo", "subnet_id", "ucloud_subnet.foo", "id"),
					resource.TestCheckResourceAttrPair("ucloud_route_table_association.foo", "route_table_id", "ucloud_route_table.foo", "id"),
				),
			},
			{
				Config: testAccRouteTableAssociationConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists("ucloud_route_table_association.foo", &subnet),
					testAccCheckRouteTableAssociationAttributes(&subnet),
					resource.TestCheckResourceAttrPair("ucloud_route_table_association.foo", "subnet_id", "ucloud_subnet.foo", "id"),
					resource.TestCheckResourceAttrPair("ucloud_route_table_association.foo", "route_table_id", "ucloud_route_table.bar", "id"),
				),
			},
		},
	})
}

const testAccRouteTableAssociationConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-route-table-association"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-route-table-association"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_route_table" "foo" {
	name   = "tf-acc-route-table-association-foo"
	tag    = "tf-acc"
	vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table" "bar" {
	name   = "tf-acc-route-table-association-bar"
	tag    = "tf-acc"
	vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table_association" "foo" {
	subnet_id     = ucloud_subnet.foo.id
	route_table_id = ucloud_route_table.foo.id
}
`

const testAccRouteTableAssociationConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-route-table-association"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
	name       = "tf-acc-route-table-association"
	tag        = "tf-acc"
	cidr_block = "192.168.1.0/24"
	vpc_id     = ucloud_vpc.foo.id
}

resource "ucloud_route_table" "foo" {
	name   = "tf-acc-route-table-association-foo"
	tag    = "tf-acc"
	vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table" "bar" {
	name   = "tf-acc-route-table-association-bar"
	tag    = "tf-acc"
	vpc_id = ucloud_vpc.foo.id
}

resource "ucloud_route_table_association" "foo" {
	subnet_id     = ucloud_subnet.foo.id
	route_table_id = ucloud_route_table.bar.id
}
`
