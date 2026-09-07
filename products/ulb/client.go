package ulb

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

// productClient contains the SDK clients used by the legacy ULB surface.
// Both clients share the provider runtime's configuration and credentials.
type productClient struct {
	ulbconn  *ulb.ULBClient
	unetconn *unet.UNetClient
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
	client := &productClient{
		ulbconn:  ulb.NewClient(config, credential),
		unetconn: unet.NewClient(config, credential),
	}
	for _, handler := range handlers {
		_ = client.ulbconn.AddHttpRequestHandler(handler)
		_ = client.unetconn.AddHttpRequestHandler(handler)
	}
	return client
}
