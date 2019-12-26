package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func TestAccUCloudLB_basic(t *testing.T) {
	var lbSet ulb.ULBSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_lb.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckLBDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccLBConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBAttributes(&lbSet),
					resource.TestCheckResourceAttr("ucloud_lb.foo", "name", "tf-acc-lb"),
					resource.TestCheckResourceAttr("ucloud_lb.foo", "tag", "tf-acc"),
				),
			},

			{
				Config: testAccLBConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBExists("ucloud_lb.foo", &lbSet),
					testAccCheckLBAttributes(&lbSet),
					resource.TestCheckResourceAttr("ucloud_lb.foo", "name", "tf-acc-lb-two"),
					resource.TestCheckResourceAttr("ucloud_lb.foo", "tag", defaultTag),
				),
			},
		},
	})

}

func testAccCheckLBExists(n string, lbSet *ulb.ULBSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("lb id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeLBById(rs.Primary.ID)

		log.Printf("[INFO] lb id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*lbSet = *ptr
		return nil
	}
}

func testAccCheckLBAttributes(lbSet *ulb.ULBSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if lbSet.ULBId == "" {
			return fmt.Errorf("lb id is empty")
		}
		return nil
	}
}

func testAccCheckLBDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_lb" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeLBById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.ULBId != "" {
			return fmt.Errorf("LB still exist")
		}
	}

	return nil
}

const testAccLBConfig = `
resource "ucloud_lb" "foo" {
	name = "tf-acc-lb"
	tag  = "tf-acc"
}
`
const testAccLBConfigTwo = `
resource "ucloud_lb" "foo" {
	name = "tf-acc-lb-two"
	tag  = ""
}
`
