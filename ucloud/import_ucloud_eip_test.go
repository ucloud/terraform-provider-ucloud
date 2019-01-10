package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudEIP_import(t *testing.T) {
	resourceName := "ucloud_eip.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"duration"},
			},
		},
	})
}
