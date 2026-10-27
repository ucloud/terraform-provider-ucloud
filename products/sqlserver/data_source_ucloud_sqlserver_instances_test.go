package sqlserver_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudSQLServerInstances_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSQLServerInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSQLServerInstanceConfig + testAccSQLServerInstancesDataSourceConfig,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ucloud_sqlserver_instances.default", "total_count"),
				),
			},
		},
	})
}

const testAccSQLServerInstancesDataSourceConfig = `
data "ucloud_sqlserver_instances" "default" {
  availability_zone = data.ucloud_zones.default.zones[0].id
}
`
