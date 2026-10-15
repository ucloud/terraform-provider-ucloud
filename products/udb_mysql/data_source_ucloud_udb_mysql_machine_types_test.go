package udb_mysql_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudMySQLMachineTypes_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: `
data "ucloud_zones" "default" {}

data "ucloud_udb_mysql_machine_types" "default" {
  availability_zone = data.ucloud_zones.default.zones[0].id
}
`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ucloud_udb_mysql_machine_types.default", "total_count"),
				),
			},
		},
	})
}
