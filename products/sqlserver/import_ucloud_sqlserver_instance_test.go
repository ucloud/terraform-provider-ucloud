package sqlserver_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudSQLServerInstance_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSQLServerInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSQLServerInstanceConfig,
			},
			{
				ResourceName:            "ucloud_sqlserver_instance.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
