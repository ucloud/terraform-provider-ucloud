package iam

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"

	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// productClient contains the SDK client used by IAM resources and data sources.
type productClient struct {
	iamconn *iam.IAMClient
}

func clientFromMeta(meta interface{}) (*productClient, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*productClient)
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
	client := &productClient{iamconn: iam.NewClient(config, credential)}
	for _, handler := range handlers {
		client.iamconn.AddHttpRequestHandler(handler)
	}
	return client
}
