package umongodb

import (
	"fmt"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func clientFromMeta(meta interface{}) (*umongodb.UMongoDBClient, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*umongodb.UMongoDBClient)
	if !ok {
		return nil, fmt.Errorf("product client %q has unexpected type %T", Name, client)
	}
	return typed, nil
}

func newClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longtimeConfig := *config
	longtimeConfig.Timeout = 60 * time.Second
	client := umongodb.NewClient(&longtimeConfig, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

func providerRegion(client *umongodb.UMongoDBClient) string {
	if client == nil || client.GetConfig() == nil {
		return ""
	}
	return client.GetConfig().Region
}
