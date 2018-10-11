package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudEIP_basic(t *testing.T) {
	var eip unet.UnetEIPSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_eip.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccEIPConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAttributes(&eip),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "bandwidth", "1"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "name", "testAcc"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "internet_charge_mode", "Bandwidth"),
				),
			},

			resource.TestStep{
				Config: testAccEIPConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAttributes(&eip),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "internet_charge_mode", "Traffic"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "bandwidth", "2"),
					resource.TestCheckResourceAttr("ucloud_eip.foo", "name", "testAccTwo"),
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
	name = "testAcc"
	bandwidth = 1
	internet_charge_mode = "Bandwidth"
}
`
const testAccEIPConfigTwo = `
resource "ucloud_eip" "foo" {
	name = "testAccTwo"
	bandwidth = 2
	internet_charge_mode = "Traffic"
}
`
