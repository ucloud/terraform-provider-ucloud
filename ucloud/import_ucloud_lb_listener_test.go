package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudLBListener_import(t *testing.T) {
	resourceName := "ucloud_lb_listener.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBListenerConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
