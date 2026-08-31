package uphost

import (
	"fmt"
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	sdkunet "github.com/ucloud/ucloud-sdk-go/services/unet"
	sdkuphost "github.com/ucloud/ucloud-sdk-go/services/uphost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// productClient contains only the SDK clients used by the legacy UPHost
// resource and data source. Region and project ID are retained locally to
// preserve the legacy client identity alongside the SDK configuration.
type productClient struct {
	region    string
	projectId string

	uphostconn    *sdkuphost.UPHostClient
	unetconn      *sdkunet.UNetClient
	genericClient *ucloud.Client
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

	// Keep the legacy constructor split: UNet uses the provider config while
	// UPHost and generic invocation use the long-timeout config.
	unetconn := sdkunet.NewClient(config, credential)
	uphostconn := sdkuphost.NewClient(&longTimeoutConfig, credential)
	genericClient := ucloud.NewClient(&longTimeoutConfig, credential)
	client := &productClient{
		region:        config.Region,
		projectId:     config.ProjectId,
		uphostconn:    uphostconn,
		unetconn:      unetconn,
		genericClient: genericClient,
	}
	for _, handler := range handlers {
		_ = client.unetconn.AddHttpRequestHandler(handler)
		_ = client.uphostconn.AddHttpRequestHandler(handler)
		_ = client.genericClient.AddHttpRequestHandler(handler)
	}
	return client
}
