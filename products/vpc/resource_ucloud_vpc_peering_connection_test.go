package vpc_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func TestAccUCloudVPCPeeringConnection_basic(t *testing.T) {
	var vpc1 vpcapi.VPCInfo
	var vpc2 vpcapi.VPCInfo
	var val vpcapi.VPCIntercomInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_vpc_peering_connection.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVPCPeeringConnectionDestroy,

		Steps: []resource.TestStep{
			{
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

func describeAccVPCIntercomByID(
	client *vpcapi.VPCClient,
	vpcID string,
	peerVPCID string,
	peerRegion string,
	peerProjectID string,
) (*vpcapi.VPCIntercomInfo, bool, error) {
	if vpcID == "" {
		return nil, false, fmt.Errorf("vpc id is empty")
	}
	request := client.NewDescribeVPCIntercomRequest()
	request.VPCId = ucloud.String(vpcID)
	request.DstRegion = ucloud.String(peerRegion)
	request.DstProjectId = ucloud.String(peerProjectID)
	response, err := client.DescribeVPCIntercom(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading vpc peer connection %q, %s", vpcID, response.GetMessage())
	}
	for index := range response.DataSet {
		if response.DataSet[index].VPCId == peerVPCID {
			return &response.DataSet[index], true, nil
		}
	}
	return nil, false, nil
}

func testAccCheckVPCPeeringConnectionExists(name string, target *vpcapi.VPCIntercomInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("vpc id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccVPCIntercomByID(
			clients.vpcconn,
			item.Primary.Attributes["vpc_id"],
			item.Primary.Attributes["peer_vpc_id"],
			clients.region,
			item.Primary.Attributes["peer_project_id"],
		)
		log.Printf("[INFO] vpc id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("vpc peer connection %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckVPCPeeringConnectionAttributes(value *vpcapi.VPCInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.VPCId == "" {
			return fmt.Errorf("vpc peering connection id is empty")
		}
		return nil
	}
}

func testAccCheckVPCPeeringConnectionDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_vpc_peering_connection" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccVPCIntercomByID(
			clients.vpcconn,
			item.Primary.Attributes["vpc_id"],
			item.Primary.Attributes["peer_vpc_id"],
			clients.region,
			item.Primary.Attributes["peer_project_id"],
		)
		if err != nil {
			return err
		}
		if found && value.VPCId != "" {
			return fmt.Errorf("VPC still exist")
		}
	}
	return nil
}

const testAccVPCPeeringConnectionConfig = `
resource "ucloud_vpc" "foo" {
	name        = "tf-acc-vpc"
	tag         = "tf-acc"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_vpc" "bar" {
	name        = "tf-acc-vpc"
	tag         = "tf-acc"
	cidr_blocks = ["10.10.0.0/16"]
}

resource "ucloud_vpc_peering_connection" "foo" {
	vpc_id      = "${ucloud_vpc.foo.id}"
	peer_vpc_id = "${ucloud_vpc.bar.id}"
}
`
