package unet_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productunet "github.com/terraform-providers/terraform-provider-ucloud/products/unet"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
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

func testAccUNetClient() (*unet.UNetClient, error) {
	client, err := testAccHarness.ProductClient(productunet.Name, newAccUNetClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*unet.UNetClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UNet acceptance client type %T", client)
	}
	return typed, nil
}

func newAccUNetClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	client := unet.NewClient(config, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeAccEIPByID(
	client *unet.UNetClient,
	id string,
) (*unet.UnetEIPSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("eip id is empty")
	}
	request := client.NewDescribeEIPRequest()
	request.EIPIds = []string{id}
	response, err := client.DescribeEIP(request)
	if err != nil {
		if isUNetNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.EIPSet) == 0 {
		return nil, false, nil
	}
	return &response.EIPSet[0], true, nil
}

func describeAccEIPResourceByID(
	client *unet.UNetClient,
	eipID string,
	resourceID string,
) (*unet.UnetEIPResourceSet, bool, error) {
	if eipID == "" {
		return nil, false, fmt.Errorf("eip id is empty")
	}
	request := client.NewDescribeEIPRequest()
	request.EIPIds = []string{eipID}
	response, err := client.DescribeEIP(request)
	if err != nil {
		if isUNetNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.EIPSet) == 0 {
		return nil, false, nil
	}
	for index := range response.EIPSet {
		binding := response.EIPSet[index].Resource
		if binding.ResourceID == resourceID || binding.ResourceId == resourceID {
			return &binding, true, nil
		}
	}
	return nil, false, nil
}

func describeAccSecurityGroupByID(
	client *unet.UNetClient,
	id string,
) (*unet.FirewallDataSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("security group id is empty")
	}
	request := client.NewDescribeFirewallRequest()
	request.FWId = ucloud.String(id)
	response, err := client.DescribeFirewall(request)
	if err != nil {
		if isUNetNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func isUNetNotFound(err error) bool {
	cloudErr, ok := err.(uerr.Error)
	return ok && cloudErr.Code() == 54002
}

func testAccCheckEIPDestroy(state *terraform.State) error {
	client, err := testAccUNetClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_eip" {
			continue
		}
		eip, found, err := describeAccEIPByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if eip.EIPId != "" {
			return fmt.Errorf("EIP still exist")
		}
	}
	return nil
}

func testAccCheckEIPAssociationDestroy(state *terraform.State) error {
	client, err := testAccUNetClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_eip_association" {
			continue
		}
		eipID := item.Primary.Attributes["eip_id"]
		resourceID := item.Primary.Attributes["resource_id"]
		binding, found, err := describeAccEIPResourceByID(client, eipID, resourceID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if binding.ResourceID == resourceID || binding.ResourceId == resourceID {
			return fmt.Errorf("eip association still exists")
		}
	}
	return nil
}

func testAccCheckSecurityGroupDestroy(state *terraform.State) error {
	client, err := testAccUNetClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_security_group" {
			continue
		}
		securityGroup, found, err := describeAccSecurityGroupByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if securityGroup.FWId != "" {
			return fmt.Errorf("security group still exist")
		}
	}
	return nil
}
