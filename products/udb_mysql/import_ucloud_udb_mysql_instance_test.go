package udb_mysql_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudMySQLInstance_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMySQLInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccMySQLInstanceConfig,
			},
			{
				ResourceName:            "ucloud_udb_mysql_instance.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
