package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudRouteTable_basic(t *testing.T) {
	var val vpcapi.RouteTableInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_route_table.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists("ucloud_route_table.foo", &val),
					testAccCheckRouteTableAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_route_table.foo", "name", "tf-acc-route-table"),
					resource.TestCheckResourceAttr("ucloud_route_table.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_route_table.foo", "remark", "tf-acc-route-table"),
				),
			},
			{
				Config: testAccRouteTableConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists("ucloud_route_table.foo", &val),
					testAccCheckRouteTableAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_route_table.foo", "name", "tf-acc-route-table-updated"),
					resource.TestCheckResourceAttr("ucloud_route_table.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_route_table.foo", "remark", "tf-acc-route-table-updated"),
				),
			},
		},
	})
}

const testAccRouteTableConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpc"
	tag         = ""
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_route_table" "foo" {
	name    = "tf-acc-route-table"
	tag     = ""
	remark  = "tf-acc-route-table"
	vpc_id  = ucloud_vpc.foo.id
}
`

const testAccRouteTableConfigUpdate = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpc"
	tag         = ""
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_route_table" "foo" {
	name    = "tf-acc-route-table-updated"
	tag     = ""
	remark  = "tf-acc-route-table-updated"
	vpc_id  = ucloud_vpc.foo.id
}
`
