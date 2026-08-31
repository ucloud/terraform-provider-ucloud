package udisk_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	udiskapi "github.com/ucloud/ucloud-sdk-go/services/udisk"
	uhostapi "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productudisk "github.com/terraform-providers/terraform-provider-ucloud/products/udisk"
)

const (
	defaultTag           = "Default"
	snapshotStatusNormal = "Normal"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

type acceptanceClients struct {
	udisk *udiskapi.UDiskClient
	uhost *uhostapi.UHostClient
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccClients() (*acceptanceClients, error) {
	client, err := testAccHarness.ProductClient(productudisk.Name, newAcceptanceClients)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*acceptanceClients)
	if !ok {
		return nil, fmt.Errorf("unexpected UDisk acceptance client type %T", client)
	}
	return typed, nil
}

func newAcceptanceClients(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	clients := &acceptanceClients{
		udisk: udiskapi.NewClient(&longTimeoutConfig, credential),
		uhost: uhostapi.NewClient(&longTimeoutConfig, credential),
	}
	for _, handler := range handlers {
		clients.udisk.AddHttpRequestHandler(handler)
		clients.uhost.AddHttpRequestHandler(handler)
	}
	return clients
}

func describeAccDiskByID(client *udiskapi.UDiskClient, id string) (*udiskapi.UDiskDataSet, bool, error) {
	request := client.NewDescribeUDiskRequest()
	request.UDiskId = ucloud.String(id)
	response, err := client.DescribeUDisk(request)
	if err != nil {
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading disk %q, %s", id, response.GetMessage())
	}
	if len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccDiskSnapshotByID(client *udiskapi.UDiskClient, id string) (*udiskapi.UDiskSnapshotSet, bool, error) {
	request := client.NewDescribeUDiskSnapshotRequest()
	request.SnapshotId = ucloud.String(id)
	response, err := client.DescribeUDiskSnapshot(request)
	if err != nil {
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading disk snapshot %q, %s", id, response.GetMessage())
	}
	if len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccDiskResource(client *udiskapi.UDiskClient, diskID, instanceID string) (*udiskapi.UDiskDataSet, bool, error) {
	disk, found, err := describeAccDiskByID(client, diskID)
	if err != nil || !found {
		return disk, found, err
	}
	if disk.UHostId != instanceID {
		return nil, false, nil
	}
	return disk, true, nil
}

func testAccCheckInstanceExists(name string, instance *uhostapi.UHostInstanceSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if resourceState.Primary.ID == "" {
			return fmt.Errorf("instance id is empty")
		}

		clients, err := testAccClients()
		if err != nil {
			return err
		}
		request := clients.uhost.NewDescribeUHostInstanceRequest()
		request.UHostIds = []string{resourceState.Primary.ID}
		response, err := clients.uhost.DescribeUHostInstance(request)
		log.Printf("[INFO] instance id %#v", resourceState.Primary.ID)
		if err != nil {
			return err
		}
		if response == nil || response.GetRetCode() != 0 || len(response.UHostSet) == 0 {
			if response == nil {
				return fmt.Errorf("instance %q is not found", resourceState.Primary.ID)
			}
			return fmt.Errorf("instance %q is not found: %s", resourceState.Primary.ID, response.GetMessage())
		}
		*instance = response.UHostSet[0]
		return nil
	}
}
