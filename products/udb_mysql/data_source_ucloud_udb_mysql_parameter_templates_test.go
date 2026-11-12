package udb_mysql_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudMySQLParameterTemplates_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: `
data "ucloud_udb_mysql_parameter_templates" "default" {
  db_version = "mysql-8.0"
}
`,

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.ucloud_udb_mysql_parameter_templates.default", "total_count"),
				),
			},
		},
	})
}
