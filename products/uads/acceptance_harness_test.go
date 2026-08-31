package uads_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productuads "github.com/terraform-providers/terraform-provider-ucloud/products/uads"
	sdkuads "github.com/ucloud/ucloud-sdk-go/services/uads"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccUADSClient() (*sdkuads.UADSClient, error) {
	client, err := testAccHarness.ProductClient(productuads.Name, newAccUADSClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*sdkuads.UADSClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UADS acceptance client type %T", client)
	}
	return typed, nil
}

func newAccUADSClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	client := sdkuads.NewClient(config, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeAccUADSByID(
	client *sdkuads.UADSClient,
	id string,
) (*sdkuads.ServiceInfo, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("uads id is empty")
	}
	request := client.NewDescribeNapServiceInfoRequest()
	request.ResourceId = ucloud.String(id)
	request.NapType = ucloud.Int(1)
	request.ProjectId = nil
	response, err := client.DescribeNapServiceInfo(request)
	if err != nil {
		return nil, false, err
	}
	if response != nil && response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading uads %q, %s", id, response.GetMessage())
	}
	if response == nil || len(response.ServiceInfo) == 0 {
		return nil, false, nil
	}
	return &response.ServiceInfo[0], true, nil
}

func testAccCheckAntiDDoSInstanceExists(
	name string,
	uadsServiceInfo *sdkuads.ServiceInfo,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("uads id is empty")
		}

		client, err := testAccUADSClient()
		if err != nil {
			return err
		}
		remote, found, err := describeAccUADSByID(client, item.Primary.ID)

		log.Printf("[INFO] disk id %#v", item.Primary.ID)

		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("uads %q is not found", item.Primary.ID)
		}

		*uadsServiceInfo = *remote
		return nil
	}
}

func testAccCheckAntiDDoSInstanceAttributes(uadsServiceInfo *sdkuads.ServiceInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if uadsServiceInfo.ResourceId == "" {
			return fmt.Errorf("uads id is empty")
		}
		return nil
	}
}

func testAccCheckAntiDDoSInstanceDestroy(state *terraform.State) error {
	client, err := testAccUADSClient()
	if err != nil {
		return err
	}

	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_anti_ddos_instance" {
			continue
		}

		remote, found, err := describeAccUADSByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}

		if remote.ResourceId != "" {
			return fmt.Errorf("ucloud_anti_ddos_instance still exist")
		}
	}

	return nil
}
