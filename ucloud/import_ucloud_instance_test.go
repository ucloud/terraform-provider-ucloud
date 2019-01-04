package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudInstance_import(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "ucloud_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigVPC(rInt),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,

				ImportStateVerifyIgnore: []string{
					"root_password",
					"security_group",
					"duration",
				},
			},
		},
	})
}
