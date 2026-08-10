package ucloud

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (client *UCloudClient) describeDiskById(diskId string) (*udisk.UDiskDataSet, error) {
	if diskId == "" {
		return nil, newNotFoundError(getNotFoundMessage("disk", diskId))
	}
	req := client.udiskconn.NewDescribeUDiskRequest()
	req.UDiskId = ucloud.String(diskId)

	resp, err := client.udiskconn.DescribeUDisk(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading disk %q, %s", diskId, resp.GetMessage())
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("disk", diskId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) describeSnapshotById(snapshotId string) (*udisk.UDiskSnapshotSet, error) {
	if snapshotId == "" {
		return nil, newNotFoundError(getNotFoundMessage("disk snapshot", snapshotId))
	}
	req := client.udiskconn.NewDescribeUDiskSnapshotRequest()
	req.SnapshotId = ucloud.String(snapshotId)

	resp, err := client.udiskconn.DescribeUDiskSnapshot(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading disk snapshot %q, %s", snapshotId, resp.GetMessage())
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("disk snapshot", snapshotId))
	}

	return &resp.DataSet[0], nil
}

// describeSnapshots retrieves all the snapshots of the given zone, the diskId is optional
// and requires the zone to be specified when it is used as filter
func (client *UCloudClient) describeSnapshots(zone, diskId string) ([]udisk.UDiskSnapshotSet, error) {
	var snapshots []udisk.UDiskSnapshotSet
	limit := 100

	for offset := 0; ; offset += limit {
		req := client.udiskconn.NewDescribeUDiskSnapshotRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		if zone != "" {
			req.Zone = ucloud.String(zone)
		}
		if diskId != "" {
			req.UDiskId = ucloud.String(diskId)
		}

		resp, err := client.udiskconn.DescribeUDiskSnapshot(req)
		if err != nil {
			return nil, err
		}
		if resp.GetRetCode() != 0 {
			return nil, fmt.Errorf("error on reading disk snapshot list, %s", resp.GetMessage())
		}
		if len(resp.DataSet) < 1 {
			break
		}

		snapshots = append(snapshots, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}
	}

	return snapshots, nil
}

func (client *UCloudClient) describeDiskResource(diskId, resourceId string) (*udisk.UDiskDataSet, error) {
	req := client.udiskconn.NewDescribeUDiskRequest()
	req.UDiskId = ucloud.String(diskId)

	resp, err := client.udiskconn.DescribeUDisk(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading disk_attachment %q, %s", diskId, resp.GetMessage())
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
