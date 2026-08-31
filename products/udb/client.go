package udb

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func clientFromMeta(meta interface{}) (*udb.UDBClient, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*udb.UDBClient)
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
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := udb.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}
