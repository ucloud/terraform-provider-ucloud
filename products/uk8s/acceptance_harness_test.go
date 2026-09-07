package uk8s_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productuk8s "github.com/terraform-providers/terraform-provider-ucloud/products/uk8s"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccUK8SClient() (*uk8s.UK8SClient, error) {
	client, err := testAccHarness.ProductClient(productuk8s.Name, newAccUK8SClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*uk8s.UK8SClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UK8S acceptance client type %T", client)
	}
	return typed, nil
}

func newAccUK8SClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := uk8s.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeAccUK8SClusterByID(client *uk8s.UK8SClient, id string) (*uk8s.ClusterSet, bool, error) {
	if id == "" {
		return nil, false, nil
	}
	request := client.NewListUK8SClusterV2Request()
	request.ClusterId = ucloud.String(id)
	response, err := client.ListUK8SClusterV2(request)
	if err != nil {
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading uk8s_cluster %q, %s", id, response.GetMessage())
	}
	if len(response.ClusterSet) == 0 {
		return nil, false, nil
	}
	return &response.ClusterSet[0], true, nil
}

func describeAccUK8SNodeByResourceID(client *uk8s.UK8SClient, clusterID, resourceID string) (*uk8s.NodeInfoV2, bool, error) {
	if clusterID == "" || resourceID == "" {
		return nil, false, nil
	}
	request := client.NewListUK8SClusterNodeV2Request()
	request.ClusterId = ucloud.String(clusterID)
	response, err := client.ListUK8SClusterNodeV2(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 94007 {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	for index := range response.NodeSet {
		if response.NodeSet[index].NodeId == resourceID {
			return &response.NodeSet[index], true, nil
		}
	}
	return nil, false, nil
}
