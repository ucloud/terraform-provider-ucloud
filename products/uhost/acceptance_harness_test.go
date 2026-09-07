package uhost_test

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	uhostapi "github.com/ucloud/ucloud-sdk-go/services/uhost"
	unetapi "github.com/ucloud/ucloud-sdk-go/services/unet"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	ucloudsdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productuhost "github.com/terraform-providers/terraform-provider-ucloud/products/uhost"
)

const defaultTag = "Default"

var testAccHarness = acceptancetest.New()

var testAccProvider = testAccHarness.Provider

var testAccProviders = testAccHarness.Providers

type acceptanceClients struct {
	uhost *uhostapi.UHostClient
	unet  *unetapi.UNetClient
	vpc   *vpcapi.VPCClient
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccClients() (*acceptanceClients, error) {
	client, err := testAccHarness.ProductClient(productuhost.Name, newAcceptanceClients)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*acceptanceClients)
	if !ok {
		return nil, fmt.Errorf("unexpected UHost acceptance client type %T", client)
	}
	return typed, nil
}

func newAcceptanceClients(
	config *ucloudsdk.Config,
	credential *auth.Credential,
	handlers []ucloudsdk.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	clients := &acceptanceClients{
		uhost: uhostapi.NewClient(&longTimeoutConfig, credential),
		unet:  unetapi.NewClient(config, credential),
		vpc:   vpcapi.NewClient(config, credential),
	}
	for _, handler := range handlers {
		_ = clients.uhost.AddHttpRequestHandler(handler)
		_ = clients.unet.AddHttpRequestHandler(handler)
		_ = clients.vpc.AddHttpRequestHandler(handler)
	}
	return clients
}

func isAcceptanceNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 8037 {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "notfound") || strings.Contains(message, "not found")
}

func describeAccInstanceByID(client *uhostapi.UHostClient, id string) (*uhostapi.UHostInstanceSet, bool, error) {
	request := client.NewDescribeUHostInstanceRequest()
	request.UHostIds = []string{id}
	response, err := client.DescribeUHostInstance(request)
	if err != nil {
		if isAcceptanceNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isAcceptanceNotFoundError(fmt.Errorf("%s", response.GetMessage())) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading instance %q, %s", id, response.GetMessage())
	}
	if len(response.UHostSet) == 0 {
		return nil, false, nil
	}
	return &response.UHostSet[0], true, nil
}

func testAccCheckInstanceExists(name string, instance *uhostapi.UHostInstanceSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("instance id is empty")
		}

		clients, err := testAccClients()
		if err != nil {
			return err
		}
		log.Printf("[INFO] instance id %#v", item.Primary.ID)
		remote, found, err := describeAccInstanceByID(clients.uhost, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("instance %q is not found", item.Primary.ID)
		}
		*instance = *remote
		return nil
	}
}

func testAccCheckInstanceDestroy(state *terraform.State) error {
	clients, err := testAccClients()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_instance" {
			continue
		}

		instance, found, err := describeAccInstanceByID(clients.uhost, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if instance.State != "" && instance.State != "Stopped" {
			return fmt.Errorf("found unstopped instance: %s", instance.UHostId)
		}
		if instance.UHostId != "" {
			return fmt.Errorf("instance still exist")
		}
	}
	return nil
}

func testAccCheckInstanceStateDestroy(state *terraform.State) error {
	return testAccCheckInstanceDestroy(state)
}

func describeAccIsolationGroupByID(client *uhostapi.UHostClient, id string) (*uhostapi.IsolationGroup, bool, error) {
	request := client.NewDescribeIsolationGroupRequest()
	request.GroupId = ucloudsdk.String(id)
	response, err := client.DescribeIsolationGroup(request)
	if err != nil {
		if isAcceptanceNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.IsolationGroupSet) == 0 {
		return nil, false, nil
	}
	return &response.IsolationGroupSet[0], true, nil
}

func testAccCheckIsolationGroupExists(name string, group *uhostapi.IsolationGroup) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("isolation group id is empty")
		}

		clients, err := testAccClients()
		if err != nil {
			return err
		}
		log.Printf("[INFO] isolation group id %#v", item.Primary.ID)
		remote, found, err := describeAccIsolationGroupByID(clients.uhost, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("isolation group %q is not found", item.Primary.ID)
		}
		*group = *remote
		return nil
	}
}

func testAccCheckIsolationGroupAttributes(group *uhostapi.IsolationGroup) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if group.GroupId == "" {
			return fmt.Errorf("isolation group id is empty")
		}
		return nil
	}
}

func testAccCheckIsolationGroupDestroy(state *terraform.State) error {
	clients, err := testAccClients()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_isolation_group" {
			continue
		}

		group, found, err := describeAccIsolationGroupByID(clients.uhost, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if group.GroupId != "" {
			return fmt.Errorf("isolation group still exist")
		}
	}
	return nil
}
