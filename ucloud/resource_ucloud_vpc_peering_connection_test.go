package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

func TestAccUCloudVPCPeeringConnection_basic(t *testing.T) {
	var vpc1 vpc.VPCInfo
	var vpc2 vpc.VPCInfo
	var val vpc.VPCIntercomInfo

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpc_peering_connection.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPCPeeringConnectionDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVPCPeeringConnectionConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCExists("ucloud_vpc.foo", &vpc1),
					testAccCheckVPCExists("ucloud_vpc.bar", &vpc2),
					testAccCheckVPCPeeringConnectionExists("ucloud_vpc_peering_connection.foo", &val),
					testAccCheckVPCAttributes(&vpc1),
					testAccCheckVPCAttributes(&vpc2),
				),
			},
		},
	})
}

func testAccCheckVPCPeeringConnectionExists(n string, val *vpc.VPCIntercomInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("vpc id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeVPCIntercomById(
			rs.Primary.Attributes["vpc_id"],
			rs.Primary.Attributes["peer_vpc_id"],
			client.region,
			rs.Primary.Attributes["peer_project_id"],
		)

		log.Printf("[INFO] vpc id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*val = *ptr
		return nil
	}
}

func testAccCheckVPCPeeringConnectionAttributes(val *vpc.VPCInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if val.VPCId == "" {
			return fmt.Errorf("vpc peering connection id is empty")
		}

		return nil
	}
}

func testAccCheckVPCPeeringConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_vpc_peering_connection" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeVPCIntercomById(
			rs.Primary.Attributes["vpc_id"],
			rs.Primary.Attributes["peer_vpc_id"],
			client.region,
			rs.Primary.Attributes["peer_project_id"],
		)

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

const testAccVPCPeeringConnectionConfig = `
resource "ucloud_vpc" "foo" {
	name = "testAcc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_vpc" "bar" {
	name = "testAcc"
	cidr_blocks = ["10.10.0.0/16"]
}

resource "ucloud_vpc_peering_connection" "foo" {
	vpc_id = "${ucloud_vpc.foo.id}"
	peer_vpc_id = "${ucloud_vpc.bar.id}"
}
`
