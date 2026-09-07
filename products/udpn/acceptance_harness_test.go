package udpn_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	udpnapi "github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productudpn "github.com/terraform-providers/terraform-provider-ucloud/products/udpn"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func TestMain(main *testing.M) {
	resource.TestMain(main)
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccUDPNClient() (*udpnapi.UDPNClient, error) {
	client, err := testAccHarness.ProductClient(productudpn.Name, newAccUDPNClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*udpnapi.UDPNClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UDPN acceptance client type %T", client)
	}
	return typed, nil
}

func testAccUDPNClientForRegion(region string) (*udpnapi.UDPNClient, error) {
	for _, name := range []string{"UCLOUD_PUBLIC_KEY", "UCLOUD_PRIVATE_KEY", "UCLOUD_PROJECT_ID"} {
		if os.Getenv(name) == "" {
			return nil, fmt.Errorf("%s must be set for the UDPN sweeper", name)
		}
	}
	config := ucloud.NewConfig()
	config.Region = region
	config.ProjectId = os.Getenv("UCLOUD_PROJECT_ID")
	credential := auth.NewCredential()
	credential.PublicKey = os.Getenv("UCLOUD_PUBLIC_KEY")
	credential.PrivateKey = os.Getenv("UCLOUD_PRIVATE_KEY")
	return newAccUDPNClient(&config, &credential, nil).(*udpnapi.UDPNClient), nil
}

func newAccUDPNClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := udpnapi.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeAccDPNByID(client *udpnapi.UDPNClient, id string) (*udpnapi.UDPNData, bool, error) {
	request := client.NewDescribeUDPNRequest()
	request.UDPNId = ucloud.String(id)
	response, err := client.DescribeUDPN(request)
	if err != nil {
		return nil, false, err
	}
	if response != nil && response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading dpn %q, %s", id, response.GetMessage())
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}
