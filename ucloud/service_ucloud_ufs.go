package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (client *UCloudClient) describeUFSById(instanceId string) (*ufs.UFSVolumeInfo2, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("ufs", instanceId))
	}
	req := client.ufsconn.NewDescribeUFSVolume2Request()
	req.VolumeId = ucloud.String(instanceId)

	resp, err := client.ufsconn.DescribeUFSVolume2(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("ufs", instanceId))
	}

	return &resp.DataSet[0], nil
}
