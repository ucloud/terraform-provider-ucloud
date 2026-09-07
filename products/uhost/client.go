package uhost

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"

	sdkuhost "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdkunet "github.com/ucloud/ucloud-sdk-go/services/unet"
	sdkvpc "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// productClient contains only the SDK clients needed by UHost resources.
// The provider runtime owns credentials, common configuration, and caching.
type productClient struct {
	uhostconn *sdkuhost.UHostClient
	unetconn  *sdkunet.UNetClient
	vpcconn   *sdkvpc.VPCClient
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
		uhostconn: sdkuhost.NewClient(&longTimeoutConfig, credential),
		unetconn:  sdkunet.NewClient(config, credential),
		vpcconn:   sdkvpc.NewClient(config, credential),
	}
	for _, handler := range handlers {
		client.uhostconn.AddHttpRequestHandler(handler)
		client.unetconn.AddHttpRequestHandler(handler)
		client.vpcconn.AddHttpRequestHandler(handler)
	}
	return client
}
