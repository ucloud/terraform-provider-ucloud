package vpc

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// productClient contains the SDK clients used by the VPC product. Region and
// project ID are kept locally because peering IDs encode both values.
type productClient struct {
	region    string
	projectId string

	vpcconn  *vpcapi.VPCClient
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
		region:    config.Region,
		projectId: config.ProjectId,
		vpcconn:   vpcapi.NewClient(config, credential),
		unetconn:  unet.NewClient(config, credential),
	}
	for _, handler := range handlers {
		_ = client.vpcconn.AddHttpRequestHandler(handler)
		_ = client.unetconn.AddHttpRequestHandler(handler)
	}
	return client
}
