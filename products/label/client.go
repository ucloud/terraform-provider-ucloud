package label

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func clientFromMeta(meta interface{}) (*labelapi.LabelClient, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*labelapi.LabelClient)
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
	client := labelapi.NewClient(config, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}
