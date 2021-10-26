package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *UCloudClient) describeUFSVolumeById(instanceId string) (*ufs.UFSVolumeInfo2, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("ufs_volume", instanceId))
	}
	req := client.ufsconn.NewDescribeUFSVolume2Request()
	req.VolumeId = ucloud.String(instanceId)

	resp, err := client.ufsconn.DescribeUFSVolume2(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("ufs_volume", instanceId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) describeUFSVolumeMountPointById(volumeId, vpcId, subnetId string) (*ufs.MountPointInfo, error) {
	req := client.ufsconn.NewDescribeUFSVolumeMountpointRequest()
	req.VolumeId = ucloud.String(volumeId)

	resp, err := client.ufsconn.DescribeUFSVolumeMountpoint(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 65126 {
			return nil, newNotFoundError(getNotFoundMessage("ufs_volume_mount_point", volumeId))
		}
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("ufs_volume_mount_point", volumeId))
	}

	for i := 0; i < len(resp.DataSet); i++ {
		resourceSet := resp.DataSet[i]
		if resourceSet.VpcId == vpcId && resourceSet.SubnetId == subnetId {
			return &resourceSet, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("ufs_volume_mount_point", volumeId))
}
