package ucloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudUDPNConnection_import(t *testing.T) {
	resourceName := "ucloud_udpn_connection.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUDPNConnectionImportConfig,
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

const testAccUDPNConnectionImportConfig = `
resource "ucloud_udpn_connection" "foo" {
	charge_type = "month"
	duration    = 1
	bandwidth   = 2
	peer_region = "cn-gd"
}
`
