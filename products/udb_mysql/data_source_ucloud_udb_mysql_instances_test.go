package udb_mysql_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudMySQLInstances_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMySQLInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccMySQLInstanceConfig + testAccMySQLInstancesDataSourceConfig,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ucloud_udb_mysql_instances.default", "total_count"),
				),
			},
		},
	})
}

const testAccMySQLInstancesDataSourceConfig = `
data "ucloud_udb_mysql_instances" "default" {
  availability_zone = data.ucloud_zones.default.zones[0].id
}
`
