package uk8s

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func clientFromMeta(meta interface{}) (*sdkuk8s.UK8SClient, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*sdkuk8s.UK8SClient)
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
	client := sdkuk8s.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}
