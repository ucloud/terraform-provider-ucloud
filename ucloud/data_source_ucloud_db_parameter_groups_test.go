package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudDBParameterGroupsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDBParameterGroupsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_db_parameter_groups.foo"),
					resource.TestCheckResourceAttr("data.ucloud_db_parameter_groups.foo", "parameter_groups.0.name", "mysql5.6默认配置"),
				),
			},
		},
	})
}

const testAccDataDBParameterGroupsConfig = `
data "ucloud_zones" "default" {
}

data "ucloud_db_parameter_groups" "foo" {
	availability_zone = data.ucloud_zones.default.zones[0].id
	name_regex		  = "mysql5.6默认配置"
}
`
