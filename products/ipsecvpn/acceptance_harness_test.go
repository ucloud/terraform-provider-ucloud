package ipsecvpn_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productipsecvpn "github.com/terraform-providers/terraform-provider-ucloud/products/ipsecvpn"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccIPSecVPNClient() (*ipsecvpn.IPSecVPNClient, error) {
	client, err := testAccHarness.ProductClient(productipsecvpn.Name, newAccIPSecVPNClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*ipsecvpn.IPSecVPNClient)
	if !ok {
		return nil, fmt.Errorf("unexpected IPSecVPN acceptance client type %T", client)
	}
	return typed, nil
}

func newAccIPSecVPNClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	client := ipsecvpn.NewClient(config, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeAccVPNGatewayByID(
	client *ipsecvpn.IPSecVPNClient,
	id string,
) (*ipsecvpn.VPNGatewayDataSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("vpn gateway id is empty")
	}
	request := client.NewDescribeVPNGatewayRequest()
	request.VPNGatewayIds = []string{id}
	response, err := client.DescribeVPNGateway(request)
	if err != nil {
		if isIPSecVPNNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccVPNCustomerGatewayByID(
	client *ipsecvpn.IPSecVPNClient,
	id string,
) (*ipsecvpn.RemoteVPNGatewayDataSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("vpn customer gateway id is empty")
	}
	request := client.NewDescribeRemoteVPNGatewayRequest()
	request.RemoteVPNGatewayIds = []string{id}
	response, err := client.DescribeRemoteVPNGateway(request)
	if err != nil {
		if isIPSecVPNNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccVPNConnectionByID(
	client *ipsecvpn.IPSecVPNClient,
	id string,
) (*ipsecvpn.VPNTunnelDataSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("vpn connection id is empty")
	}
	request := client.NewDescribeVPNTunnelRequest()
	request.VPNTunnelIds = []string{id}
	response, err := client.DescribeVPNTunnel(request)
	if err != nil {
		if isIPSecVPNNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func isIPSecVPNNotFound(err error) bool {
	cloudErr, ok := err.(uerr.Error)
	return ok && cloudErr.Code() == 54002
}

func testAccCheckVPNGatewayDestroy(state *terraform.State) error {
	client, err := testAccIPSecVPNClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_vpn_gateway" {
			continue
		}
		gateway, found, err := describeAccVPNGatewayByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if gateway.VPNGatewayId != "" {
			return fmt.Errorf("vpn gateway still exist")
		}
	}
	return nil
}

func testAccCheckVPNCustomerGatewayDestroy(state *terraform.State) error {
	client, err := testAccIPSecVPNClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_vpn_customer_gateway" {
			continue
		}
		gateway, found, err := describeAccVPNCustomerGatewayByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if gateway.RemoteVPNGatewayId != "" {
			return fmt.Errorf("vpn customer gateway still exist")
		}
	}
	return nil
}

func testAccCheckVPNConnectionDestroy(state *terraform.State) error {
	client, err := testAccIPSecVPNClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_vpn_connection" {
			continue
		}
		connection, found, err := describeAccVPNConnectionByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if connection.VPNTunnelId != "" {
			return fmt.Errorf("vpn connection still exist")
		}
	}
	return nil
}
