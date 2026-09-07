package ufs_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	ufsapi "github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productufs "github.com/terraform-providers/terraform-provider-ucloud/products/ufs"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccUFSClient() (*ufsapi.UFSClient, error) {
	client, err := testAccHarness.ProductClient(productufs.Name, func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		client := ufsapi.NewClient(config, credential)
		for _, handler := range handlers {
			client.AddHttpRequestHandler(handler)
		}
		return client
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*ufsapi.UFSClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UFS acceptance client type %T", client)
	}
	return typed, nil
}

func describeAccUFSVolumeByID(client *ufsapi.UFSClient, id string) (*ufsapi.UFSVolumeInfo2, bool, error) {
	request := client.NewDescribeUFSVolume2Request()
	request.VolumeId = ucloud.String(id)
	response, err := client.DescribeUFSVolume2(request)
	if err != nil {
		return nil, false, err
	}
	if response != nil && response.GetRetCode() != 0 {
		return nil, false, fmt.Errorf("error on reading ufs_volume %q, %s", id, response.GetMessage())
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccUFSMountPoint(
	client *ufsapi.UFSClient,
	volumeID string,
	vpcID string,
	subnetID string,
) (*ufsapi.MountPointInfo, bool, error) {
	request := client.NewDescribeUFSVolumeMountpointRequest()
	request.VolumeId = ucloud.String(volumeID)
	response, err := client.DescribeUFSVolumeMountpoint(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 65126 {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	for index := range response.DataSet {
		mountPoint := response.DataSet[index]
		if mountPoint.VpcId == vpcID && mountPoint.SubnetId == subnetID {
			return &mountPoint, true, nil
		}
	}
	return nil, false, nil
}
