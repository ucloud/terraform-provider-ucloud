package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (client *UCloudClient) describeDiskById(diskId string) (*udisk.UDiskDataSet, error) {
	req := client.udiskconn.NewDescribeUDiskRequest()
	req.UDiskId = ucloud.String(diskId)

	resp, err := client.udiskconn.DescribeUDisk(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("disk", diskId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) describeDiskResource(diskId, resourceId string) (*udisk.UDiskDataSet, error) {
	req := client.udiskconn.NewDescribeUDiskRequest()
	req.UDiskId = ucloud.String(diskId)

	resp, err := client.udiskconn.DescribeUDisk(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("disk_attachment", diskId))
	}

	for i := 0; i < len(resp.DataSet); i++ {
		resourceSet := resp.DataSet[i]
		if resourceSet.UHostId == resourceId {
			return &resourceSet, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("disk_attachment", diskId))
}
