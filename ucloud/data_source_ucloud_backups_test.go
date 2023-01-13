package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudDBBackupssDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDBBackupssConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ucloud_db_backups.foo", "db_backups.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_db_backups.foo", "db_backups.0.backup_name", "initial-back"),
				),
			},
		},
	})
}

func testAccDataDBBackupssConfig() string {
	return ` 
provider "ucloud" {
  region = "cn-bj2"
}


data "ucloud_db_backups" "foo" {
  availability_zone = "cn-bj2-05"
  name_regex        = "init.*"
}

output "backups" {
  value = data.ucloud_db_backups.foo.db_backups
}
`
}
