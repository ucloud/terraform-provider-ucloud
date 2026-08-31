package umem

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"

	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// productClient owns the public and private UMem SDK clients used by the
// resource implementations. The provider runtime owns their shared config,
// credentials, handlers, and cache lifetime.
type productClient struct {
	umemconn  *umem.UMemClient
	pumemconn *pumem.UMemClient
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
		umemconn:  umem.NewClient(config, credential),
		pumemconn: pumem.NewClient(config, credential),
	}
	for _, handler := range handlers {
		_ = client.umemconn.AddHttpRequestHandler(handler)
		_ = client.pumemconn.AddHttpRequestHandler(handler)
	}
	return client
}
