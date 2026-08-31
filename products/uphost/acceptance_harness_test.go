package uphost_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	unetapi "github.com/ucloud/ucloud-sdk-go/services/unet"
	uphostapi "github.com/ucloud/ucloud-sdk-go/services/uphost"
	ucloudsdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productuphost "github.com/terraform-providers/terraform-provider-ucloud/products/uphost"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

type acceptanceClients struct {
	uphostconn    *uphostapi.UPHostClient
	unetconn      *unetapi.UNetClient
	genericClient *ucloudsdk.Client
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccClients() (*acceptanceClients, error) {
	client, err := testAccHarness.ProductClient(productuphost.Name, newAcceptanceClients)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*acceptanceClients)
	if !ok {
		return nil, fmt.Errorf("unexpected UPHost acceptance client type %T", client)
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
		uphostconn:    uphostapi.NewClient(&longTimeoutConfig, credential),
		unetconn:      unetapi.NewClient(config, credential),
		genericClient: ucloudsdk.NewClient(&longTimeoutConfig, credential),
	}
	for _, handler := range handlers {
		_ = clients.uphostconn.AddHttpRequestHandler(handler)
		_ = clients.unetconn.AddHttpRequestHandler(handler)
		_ = clients.genericClient.AddHttpRequestHandler(handler)
	}
	return clients
}

func TestUPHostAcceptanceClientsUseProductionTimeouts(t *testing.T) {
	config := &ucloudsdk.Config{Region: "cn-bj2", ProjectId: "project-test", Timeout: 7 * time.Second}
	client, ok := newAcceptanceClients(config, &auth.Credential{}, nil).(*acceptanceClients)
	if !ok {
		t.Fatalf("newAcceptanceClients() returned an unexpected type")
	}
	if client.uphostconn == nil || client.unetconn == nil || client.genericClient == nil {
		t.Fatal("acceptance clients are incomplete")
	}
	if got := client.uphostconn.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UPHost timeout = %s, want 1m0s", got)
	}
	if got := client.genericClient.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("generic client timeout = %s, want 1m0s", got)
	}
	if got := client.unetconn.GetConfig().Timeout; got != config.Timeout {
		t.Fatalf("UNet timeout = %s, want caller timeout %s", got, config.Timeout)
	}
	if config.Timeout != 7*time.Second {
		t.Fatalf("newAcceptanceClients changed caller timeout to %s", config.Timeout)
	}
}

func isAcceptanceNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 16001 {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "notfound") || strings.Contains(message, "not found")
}

func describeAccBareMetalInstanceByID(
	client *uphostapi.UPHostClient,
	id string,
) (*uphostapi.PHostSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("instance id is empty")
	}
	request := client.NewDescribePHostRequest()
	request.PHostId = []string{id}
	response, err := client.DescribePHost(request)
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
	if len(response.PHostSet) == 0 {
		return nil, false, nil
	}
	return &response.PHostSet[0], true, nil
}

func testAccCheckBareMetalInstanceExists(
	name string,
	target *uphostapi.PHostSet,
) resource.TestCheckFunc {
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
		instance, found, err := describeAccBareMetalInstanceByID(clients.uphostconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("instance %q is not found", item.Primary.ID)
		}
		*target = *instance
		return nil
	}
}

func testAccCheckBareMetalInstanceDestroy(state *terraform.State) error {
	clients, err := testAccClients()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_baremetal_instance" {
			continue
		}
		instance, found, err := describeAccBareMetalInstanceByID(clients.uphostconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && instance.PHostId != "" {
			return errors.New("instance still exists")
		}
	}
	return nil
}
