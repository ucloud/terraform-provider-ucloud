package unet_test

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
					resource.TestCheckResourceAttr("ucloud_eip.foo", "tag", "Default"),
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
					resource.TestCheckResourceAttr("ucloud_eip.foo", "tag", "Default"),
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

func testAccCheckEIPExists(name string, eip *unet.UnetEIPSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("eip id is empty")
		}

		client, err := testAccUNetClient()
		if err != nil {
			return err
		}
		pointer, found, err := describeAccEIPByID(client, item.Primary.ID)
		log.Printf("[INFO] eip id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("eip %q is not found", item.Primary.ID)
		}
		*eip = *pointer
		return nil
	}
}

func testAccCheckEIPAttributes(eip *unet.UnetEIPSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if eip.EIPId == "" {
			return fmt.Errorf("eip id is empty")
		}
		return nil
	}
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
