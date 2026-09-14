package vpc_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudNetworkInterface_import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfig,
			},
			{
				ResourceName:      "ucloud_network_interface.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
