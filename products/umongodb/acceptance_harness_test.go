package umongodb_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productumongodb "github.com/terraform-providers/terraform-provider-ucloud/products/umongodb"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func TestMain(main *testing.M) {
	resource.TestMain(main)
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccUMongoDBClient() (*umongodb.UMongoDBClient, error) {
	client, err := testAccHarness.ProductClient(productumongodb.Name, newAccUMongoDBClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*umongodb.UMongoDBClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UMongoDB acceptance client type %T", client)
	}
	return typed, nil
}

func newAccUMongoDBClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := umongodb.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}
