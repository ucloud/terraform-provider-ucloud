package ulb_test

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	uhostapi "github.com/ucloud/ucloud-sdk-go/services/uhost"
	ulbapi "github.com/ucloud/ucloud-sdk-go/services/ulb"
	unetapi "github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productulb "github.com/terraform-providers/terraform-provider-ucloud/products/ulb"
)

const defaultTag = "Default"

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

type acceptanceClient struct {
	ulbconn  *ulbapi.ULBClient
	unetconn *unetapi.UNetClient
}

type providerError struct {
	errorCode string
	message   string
}

func (err *providerError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func newNotFoundError(message string) error {
	return &providerError{errorCode: "Notfound", message: message}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*providerError)
	return ok && (providerErr.errorCode == "Notfound" || strings.Contains(strings.ToLower(providerErr.message), "notfound"))
}

func TestMain(main *testing.M) {
	resource.TestMain(main)
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccULBClient() (*acceptanceClient, error) {
	client, err := testAccHarness.ProductClient(productulb.Name, func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		client := &acceptanceClient{
			ulbconn:  ulbapi.NewClient(config, credential),
			unetconn: unetapi.NewClient(config, credential),
		}
		for _, handler := range handlers {
			_ = client.ulbconn.AddHttpRequestHandler(handler)
			_ = client.unetconn.AddHttpRequestHandler(handler)
		}
		return client
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*acceptanceClient)
	if !ok {
		return nil, fmt.Errorf("unexpected ULB acceptance client type %T", client)
	}
	return typed, nil
}

func testAccUHostClient() (*uhostapi.UHostClient, error) {
	client, err := testAccHarness.ProductClient("uhost", newAcceptanceUHostClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*uhostapi.UHostClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UHost acceptance client type %T", client)
	}
	return typed, nil
}

func newAcceptanceUHostClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := uhostapi.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

func TestUHostAcceptanceClientUsesProductionTimeout(t *testing.T) {
	config := &ucloud.Config{
		Region:    "cn-bj2",
		ProjectId: "org-test",
		Timeout:   7 * time.Second,
	}
	client, ok := newAcceptanceUHostClient(config, &auth.Credential{}, nil).(*uhostapi.UHostClient)
	if !ok {
		t.Fatal("newAcceptanceUHostClient() returned an unexpected type")
	}
	if got := client.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UHost timeout = %s, want %s", got, 60*time.Second)
	}
	if config.Timeout != 7*time.Second {
		t.Fatalf("newAcceptanceUHostClient changed caller timeout to %s", config.Timeout)
	}
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

		client, err := testAccUHostClient()
		if err != nil {
			return err
		}
		request := client.NewDescribeUHostInstanceRequest()
		request.UHostIds = []string{item.Primary.ID}
		response, err := client.DescribeUHostInstance(request)
		log.Printf("[INFO] instance id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if response == nil || response.GetRetCode() != 0 {
			if response == nil {
				return newNotFoundError(getNotFoundMessage("instance", item.Primary.ID))
			}
			return fmt.Errorf("error on reading instance %q, %s", item.Primary.ID, response.GetMessage())
		}
		if len(response.UHostSet) < 1 {
			return newNotFoundError(getNotFoundMessage("instance", item.Primary.ID))
		}
		*instance = response.UHostSet[0]
		return nil
	}
}
