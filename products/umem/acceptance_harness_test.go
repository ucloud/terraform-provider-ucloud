package umem_test

import (
	"fmt"
	"testing"

	pumemapi "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	umemapi "github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productumem "github.com/terraform-providers/terraform-provider-ucloud/products/umem"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccUMemClient() (*umemapi.UMemClient, error) {
	client, err := testAccHarness.ProductClient(productumem.Name+":public", func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		client := umemapi.NewClient(config, credential)
		for _, handler := range handlers {
			_ = client.AddHttpRequestHandler(handler)
		}
		return client
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*umemapi.UMemClient)
	if !ok {
		return nil, fmt.Errorf("unexpected public UMem acceptance client type %T", client)
	}
	return typed, nil
}

func testAccPrivateUMemClient() (*pumemapi.UMemClient, error) {
	client, err := testAccHarness.ProductClient(productumem.Name+":private", func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		client := pumemapi.NewClient(config, credential)
		for _, handler := range handlers {
			_ = client.AddHttpRequestHandler(handler)
		}
		return client
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*pumemapi.UMemClient)
	if !ok {
		return nil, fmt.Errorf("unexpected private UMem acceptance client type %T", client)
	}
	return typed, nil
}

func describeAccActiveStandbyRedisByID(client *umemapi.UMemClient, id string) (*umemapi.URedisGroupSet, bool, error) {
	if id == "" {
		return nil, false, nil
	}
	request := client.NewDescribeURedisGroupRequest()
	request.GroupId = ucloud.String(id)
	response, err := client.DescribeURedisGroup(request)
	if err != nil {
		return nil, false, err
	}
	if response != nil && response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading redis %q, %s", id, response.GetMessage())
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccDistributedRedisByID(client *umemapi.UMemClient, id string) (*umemapi.UMemSpaceSet, bool, error) {
	if id == "" {
		return nil, false, nil
	}
	request := client.NewDescribeUMemSpaceRequest()
	request.SpaceId = ucloud.String(id)
	response, err := client.DescribeUMemSpace(request)
	if err != nil {
		return nil, false, err
	}
	if response != nil && response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading redis %q, %s", id, response.GetMessage())
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccActiveStandbyMemcacheByID(client *pumemapi.UMemClient, id string) (*pumemapi.UMemDataSet, bool, error) {
	if id == "" {
		return nil, false, nil
	}
	request := client.NewDescribeUMemRequest()
	request.ResourceId = ucloud.String(id)
	request.Protocol = ucloud.String("memcache")
	response, err := client.DescribeUMem(request)
	if err != nil {
		return nil, false, err
	}
	if response != nil && response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading memcache %q, %s", id, response.GetMessage())
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}
