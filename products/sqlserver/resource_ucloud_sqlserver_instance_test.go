package sqlserver_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudSQLServerInstance_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSQLServerInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSQLServerInstanceConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLServerInstanceExists("ucloud_sqlserver_instance.foo"),
					resource.TestCheckResourceAttr("ucloud_sqlserver_instance.foo", "name", "tf-acc-sqlserver-instance-basic"),
					resource.TestCheckResourceAttr("ucloud_sqlserver_instance.foo", "db_version", "sqlserver-2019"),
					resource.TestCheckResourceAttr("ucloud_sqlserver_instance.foo", "cpu", "2"),
					resource.TestCheckResourceAttr("ucloud_sqlserver_instance.foo", "memory", "4000"),
					resource.TestCheckResourceAttr("ucloud_sqlserver_instance.foo", "disk_space", "50"),
				),
			},
		},
	})
}

const testAccSQLServerInstanceConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_sqlserver_instance" "foo" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-acc-sqlserver-instance-basic"
  db_version        = "sqlserver-2019"
  cpu               = 2
  memory            = 4000
  disk_space        = 50
  password          = "2018_UClou"
}
`
