package ufs

import (
	"fmt"

	ufsapi "github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func describeUFSVolumeByID(client *ufsapi.UFSClient, instanceID string) (*ufsapi.UFSVolumeInfo2, error) {
	if instanceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("ufs_volume", instanceID))
	}

	request := client.NewDescribeUFSVolume2Request()
	request.VolumeId = ucloud.String(instanceID)
	response, err := client.DescribeUFSVolume2(request)
	if err != nil {
		return nil, err
	}
	if response.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading ufs_volume %q, %s", instanceID, response.GetMessage())
	}
	if len(response.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("ufs_volume", instanceID))
	}
	return &response.DataSet[0], nil
}

func describeUFSVolumeMountPointByID(client *ufsapi.UFSClient, volumeID, vpcID, subnetID string) (*ufsapi.MountPointInfo, error) {
	request := client.NewDescribeUFSVolumeMountpointRequest()
	request.VolumeId = ucloud.String(volumeID)
	response, err := client.DescribeUFSVolumeMountpoint(request)
	if err != nil {
		if ucloudErr, ok := err.(uerr.Error); ok && ucloudErr.Code() == 65126 {
			return nil, newNotFoundError(getNotFoundMessage("ufs_volume_mount_point", volumeID))
		}
		return nil, err
	}
	if len(response.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("ufs_volume_mount_point", volumeID))
	}

	for index := range response.DataSet {
		resourceSet := response.DataSet[index]
		if resourceSet.VpcId == vpcID && resourceSet.SubnetId == subnetID {
			return &resourceSet, nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("ufs_volume_mount_point", volumeID))
}
