package udisk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudDisk_import(t *testing.T) {
	resourceName := "ucloud_disk.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDiskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,

				// both are provider-side only, the DescribeUDisk API does not
				// return them, so they are absent from the imported state
				ImportStateVerifyIgnore: []string{
					"duration",
					"reboot_instance_for_resizing",
				},
			},
		},
	})
}
