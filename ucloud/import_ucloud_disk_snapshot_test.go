package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudDiskSnapshot_import(t *testing.T) {
	resourceName := "ucloud_disk_snapshot.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDiskSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDiskSnapshotConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
