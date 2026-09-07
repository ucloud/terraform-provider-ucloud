package udisk

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// productClient contains only the SDK clients needed by UDisk resources. The
// provider runtime owns credentials, common configuration, and the cache.
type productClient struct {
	udiskconn *udisk.UDiskClient
	uhostconn *uhost.UHostClient
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
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := &productClient{
		udiskconn: udisk.NewClient(config, credential),
		uhostconn: uhost.NewClient(&longTimeoutConfig, credential),
	}
	for _, handler := range handlers {
		client.udiskconn.AddHttpRequestHandler(handler)
		client.uhostconn.AddHttpRequestHandler(handler)
	}
	return client
}
