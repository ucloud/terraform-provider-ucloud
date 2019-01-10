package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudVPC_basic(t *testing.T) {
	var val vpc.VPCInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpc.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPCDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccVPCConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists("ucloud_vpc.foo", &val),
					testAccCheckVPCAttributes(&val),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "name", "tf-acc-vpc"),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_vpc.foo", "cidr_blocks.494140204", "192.168.0.0/16"),
				),
			},
		},
	})

}

func testAccCheckVPCExists(n string, val *vpc.VPCInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vpc id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVPCById(rs.Primary.ID)

		log.Printf("[INFO] vpc id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckVPCAttributes(val *vpc.VPCInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.VPCId == "" {
			return fmt.Errorf("vpc id is empty")
		}

		return nil
	}
}

func testAccCheckVPCDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_vpc" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVPCById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.VPCId != "" {
			return fmt.Errorf("VPC still exist")
		}
	}

	return nil
}

const testAccVPCConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpc"
	tag         = ""
	cidr_blocks = ["192.168.0.0/16"]
}
`
