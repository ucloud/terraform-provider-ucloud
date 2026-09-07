package udpn_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func init() {
	resource.AddTestSweepers("ucloud_udpn_connection", &resource.Sweeper{
		Name: "ucloud_udpn_connection",
		F:    testSweepUDPNConnections,
	})
}

func testSweepUDPNConnections(region string) error {
	client, err := testAccUDPNClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error in getting client when sweeping")
	}

	req := client.NewDescribeUDPNRequest()
	resp, err := client.DescribeUDPN(req)
	if err != nil {
		return fmt.Errorf("error in describing udpn connections when sweeping")
	}

	for _, item := range resp.DataSet {
		req := client.NewReleaseUDPNRequest()
		req.UDPNId = ucloud.String(item.UDPNId)

		// auto retry by ucloud sdk
		_, err := client.ReleaseUDPN(req)
		if err != nil {
			return fmt.Errorf("error in delete udpn connection when sweeping")
		}
	}

	return nil
}

func TestAccUCloudUDPNConnection_basic(t *testing.T) {
	var dpn udpn.UDPNData

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_udpn_connection.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUDPNConnectionDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUDPNConnectionConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUDPNConnectionExists("ucloud_udpn_connection.foo", &dpn),
					testAccCheckUDPNConnectionAttributes(&dpn),
					resource.TestCheckResourceAttr("ucloud_udpn_connection.foo", "bandwidth", "2"),
					resource.TestCheckResourceAttr("ucloud_udpn_connection.foo", "charge_type", "month"),
					resource.TestCheckResourceAttr("ucloud_udpn_connection.foo", "peer_region", "cn-sh2"),
				),
			},

			{
				Config: testAccUDPNConnectionConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUDPNConnectionExists("ucloud_udpn_connection.foo", &dpn),
					testAccCheckUDPNConnectionAttributes(&dpn),
					resource.TestCheckResourceAttr("ucloud_udpn_connection.foo", "bandwidth", "3"),
					resource.TestCheckResourceAttr("ucloud_udpn_connection.foo", "charge_type", "month"),
					resource.TestCheckResourceAttr("ucloud_udpn_connection.foo", "peer_region", "cn-sh2"),
				),
			},
		},
	})

}

func testAccCheckUDPNConnectionExists(n string, dpn *udpn.UDPNData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("dpn id is empty")
		}

		client, err := testAccUDPNClient()
		if err != nil {
			return err
		}
		ptr, found, err := describeAccDPNByID(client, rs.Primary.ID)

		log.Printf("[INFO] dpn id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("dpn %q is not found", rs.Primary.ID)
		}

		*dpn = *ptr
		return nil
	}
}

func testAccCheckUDPNConnectionAttributes(dpn *udpn.UDPNData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if dpn.UDPNId == "" {
			return fmt.Errorf("dpn id is empty")
		}
		return nil
	}
}

func testAccCheckUDPNConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_udpn_connection" {
			continue
		}

		client, err := testAccUDPNClient()
		if err != nil {
			return err
		}
		d, found, err := describeAccDPNByID(client, rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			return err
		}
		if !found {
			continue
		}

		if d.UDPNId != "" {
			return fmt.Errorf("DPN still exist")
		}
	}

	return nil
}

const testAccUDPNConnectionConfig = `
resource "ucloud_udpn_connection" "foo" {
	charge_type = "month"
	duration    = 1
	bandwidth   = 2
	peer_region = "cn-sh2"
}
`

const testAccUDPNConnectionConfigTwo = `
resource "ucloud_udpn_connection" "foo" {
	charge_type = "month"
	duration    = 1
	bandwidth   = 3
	peer_region = "cn-sh2"
}
`
