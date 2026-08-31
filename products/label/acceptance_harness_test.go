package label_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productlabel "github.com/terraform-providers/terraform-provider-ucloud/products/label"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccCheckLabelDestroy(state *terraform.State) error {
	client, err := testAccLabelClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_label" {
			continue
		}
		parts := strings.Split(item.Primary.ID, "#")
		if len(parts) != 2 {
			return fmt.Errorf("fail to parse label id %q", item.Primary.ID)
		}

		request := client.NewListLabelsRequest()
		request.Category = ucloud.String("custom")
		response, err := client.ListLabels(request)
		if err != nil {
			return err
		}
		if response == nil {
			continue
		}
		for _, label := range response.Labels {
			if label.Key == parts[0] && label.Value == parts[1] {
				return fmt.Errorf("label %q still exists", item.Primary.ID)
			}
		}
	}
	return nil
}

func testAccLabelClient() (*labelapi.LabelClient, error) {
	client, err := testAccHarness.ProductClient(productlabel.Name, func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		client := labelapi.NewClient(config, credential)
		for _, handler := range handlers {
			client.AddHttpRequestHandler(handler)
		}
		return client
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*labelapi.LabelClient)
	if !ok {
		return nil, fmt.Errorf("unexpected Label acceptance client type %T", client)
	}
	return typed, nil
}
