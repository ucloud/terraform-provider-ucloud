package ucloud

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudEIP_basic(t *testing.T) {
	var eip unet.UnetEIPSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_eip.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccEIPConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAttributes(&eip),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "bandwidth", "1"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "name", "tf-acc-eip"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "charge_mode", "bandwidth"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "share_bandwidth_package_id", ""),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "tag", defaultTag),
				),
			},

			{
				Config: testAccEIPConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAttributes(&eip),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "bandwidth", "2"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "name", "tf-acc-eip-two"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "charge_mode", "traffic"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "share_bandwidth_package_id", ""),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "tag", "tf-acc"),
				),
			},

			{
				Config: testAccEIPConfigThree,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAttributes(&eip),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "bandwidth", "2"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "name", "tf-acc-eip-three"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "charge_mode", "traffic"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "share_bandwidth_package_id", ""),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "tag", defaultTag),
				),
			},
		},
	})
}

func TestAccUCloudEIP_shareBandwidth(t *testing.T) {
	shareBandwidthID := os.Getenv("UCLOUD_SHARE_BANDWIDTH_ID")
	if shareBandwidthID == "" {
		t.Skip("UCLOUD_SHARE_BANDWIDTH_ID must be set for share bandwidth EIP acceptance tests")
	}

	var eip unet.UnetEIPSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_eip.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccEIPShareBandwidthConfig(shareBandwidthID),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAttributes(&eip),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "charge_mode", "share_bandwidth"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "share_bandwidth_package_id", shareBandwidthID),
					resource.TestCheckResourceAttrSet("ucloud_eip.foo", "bandwidth"),
				),
			},
		},
	})
}

func testAccCheckEIPExists(n string, eip *unet.UnetEIPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("eip id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeEIPById(rs.Primary.ID)

		log.Printf("[INFO] eip id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*eip = *ptr
		return nil
	}
}

func testAccCheckEIPAttributes(eip *unet.UnetEIPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if eip.EIPId == "" {
			return fmt.Errorf("eip id is empty")
		}
		return nil
	}
}

func testAccCheckEIPDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_eip" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeEIPById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.EIPId != "" {
			return fmt.Errorf("EIP still exist")
		}
	}

	return nil
}

const testAccEIPConfig = `
resource "ucloud_eip" "foo" {
	name          = "tf-acc-eip"
	bandwidth     = 1
	internet_type = "bgp"
	charge_mode   = "bandwidth"
	tag           = ""
}
`

const testAccEIPConfigTwo = `
resource "ucloud_eip" "foo" {
	name          = "tf-acc-eip-two"
	bandwidth     = 2
	internet_type = "bgp"
	charge_mode   = "traffic"
	tag           = "tf-acc"
}
`

const testAccEIPConfigThree = `
resource "ucloud_eip" "foo" {
	name          = "tf-acc-eip-three"
	bandwidth     = 2
	internet_type = "bgp"
	charge_mode   = "traffic"
	tag           = ""
}
`

func testAccEIPShareBandwidthConfig(shareBandwidthID string) string {
	return fmt.Sprintf(`
resource "ucloud_eip" "foo" {
  name                       = "tf-acc-eip-share-bandwidth"
  internet_type              = "bgp"
  charge_mode                = "share_bandwidth"
  share_bandwidth_package_id = %q
  bandwidth                  = 0
}
`, shareBandwidthID)
}
