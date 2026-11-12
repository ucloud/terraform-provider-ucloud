package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudSecGroup_basic(t *testing.T) {
	var val vpcapi.SecGroupInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_sec_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecGroupConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &val),
					testAccCheckSecGroupAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "name", "tf-acc-sec-group"),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "remark", "tf-acc-sec-group-remark"),
					resource.TestCheckResourceAttrSet("ucloud_sec_group.foo", "vpc_id"),
					resource.TestCheckResourceAttrSet("ucloud_sec_group.foo", "type"),
					resource.TestCheckResourceAttrSet("ucloud_sec_group.foo", "create_time"),
				),
			},
			{
				// Renames the group and changes its remark, so both update calls
				// are driven in one step.
				Config: testAccSecGroupConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecGroupExists("ucloud_sec_group.foo", &val),
					testAccCheckSecGroupAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "name", "tf-acc-sec-group-update"),
					resource.TestCheckResourceAttr("ucloud_sec_group.foo", "remark", "tf-acc-sec-group-remark-update"),
				),
			},
		},
	})
}

const testAccSecGroupConfigVPC = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-sec-group-vpc"
	tag         = ""
	cidr_blocks = ["192.168.0.0/16"]
}
`

const testAccSecGroupConfig = testAccSecGroupConfigVPC + `
resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group"
	vpc_id = ucloud_vpc.foo.id
	remark = "tf-acc-sec-group-remark"
}
`

const testAccSecGroupConfigUpdate = testAccSecGroupConfigVPC + `
resource "ucloud_sec_group" "foo" {
	name   = "tf-acc-sec-group-update"
	vpc_id = ucloud_vpc.foo.id
	remark = "tf-acc-sec-group-remark-update"
}
`
