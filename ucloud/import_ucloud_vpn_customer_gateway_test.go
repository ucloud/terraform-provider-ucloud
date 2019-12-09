package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudVPNCustomerGateway_import(t *testing.T) {
	resourceName := "ucloud_vpn_customer_gateway.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPNCusGWDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNCusGWConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
