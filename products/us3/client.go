package us3

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func clientFromMeta(meta interface{}) (*ufile.UFileClient, error) {
	runtime, ok := meta.(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", meta)
	}

	client, err := runtime.ProductClient(Name, newClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*ufile.UFileClient)
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
	client := ufile.NewClient(config, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeUS3BucketById(client *ufile.UFileClient, instanceID string) (*ufile.UFileBucketSet, error) {
	if instanceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("us3_bucket", instanceID))
	}
	req := client.NewDescribeBucketRequest()
	req.BucketName = ucloud.String(instanceID)

	resp, err := client.DescribeBucket(req)
	if err != nil {
		if ucloudErr, ok := err.(uerr.Error); ok && ucloudErr.Code() == 15010 {
			return nil, newNotFoundError(getNotFoundMessage("us3_bucket", instanceID))
		}
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("us3_bucket", instanceID))
	}

	return &resp.DataSet[0], nil
}
