package udb_mysql_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudMySQLInstance_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMySQLInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccMySQLInstanceConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckMySQLInstanceExists("ucloud_udb_mysql_instance.foo"),
					resource.TestCheckResourceAttr("ucloud_udb_mysql_instance.foo", "name", "tf-acc-mysql-instance-basic"),
					resource.TestCheckResourceAttr("ucloud_udb_mysql_instance.foo", "db_version", "mysql-8.0"),
					resource.TestCheckResourceAttr("ucloud_udb_mysql_instance.foo", "machine_type", "o.mysql2m.small"),
					resource.TestCheckResourceAttr("ucloud_udb_mysql_instance.foo", "disk_space", "20"),
				),
			},
			{
				Config: testAccMySQLInstanceConfigResized,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckMySQLInstanceExists("ucloud_udb_mysql_instance.foo"),
					resource.TestCheckResourceAttr("ucloud_udb_mysql_instance.foo", "machine_type", "o.mysql2m.medium"),
					resource.TestCheckResourceAttr("ucloud_udb_mysql_instance.foo", "disk_space", "30"),
				),
			},
		},
	})
}

const testAccMySQLInstanceConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_udb_mysql_instance" "foo" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-acc-mysql-instance-basic"
  db_version        = "mysql-8.0"
  machine_type      = "o.mysql2m.small"
  disk_space        = 20
  password          = "2018_UClou"
}
`

const testAccMySQLInstanceConfigResized = `
data "ucloud_zones" "default" {}

resource "ucloud_udb_mysql_instance" "foo" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-acc-mysql-instance-basic"
  db_version        = "mysql-8.0"
  machine_type      = "o.mysql2m.medium"
  disk_space        = 30
  password          = "2018_UClou"
}
`
