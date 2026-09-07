package udisk

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (client *productClient) describeDiskById(diskId string) (*udisk.UDiskDataSet, error) {
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

func (client *productClient) describeSnapshotById(snapshotId string) (*udisk.UDiskSnapshotSet, error) {
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
func (client *productClient) describeSnapshots(zone, diskId string) ([]udisk.UDiskSnapshotSet, error) {
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

func (client *productClient) describeDiskResource(diskId, resourceId string) (*udisk.UDiskDataSet, error) {
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

func (client *productClient) describeInstanceById(instanceID string) (*uhost.UHostInstanceSet, error) {
	if instanceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceID))
	}
	req := client.uhostconn.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{instanceID}

	resp, err := client.uhostconn.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading instance %q, %s", instanceID, resp.GetMessage())
	}
	if len(resp.UHostSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceID))
	}
	return &resp.UHostSet[0], nil
}

func (client *productClient) stopInstanceByID(instanceID string) error {
	if instanceID == "" {
		return newNotFoundError(getNotFoundMessage("instance", instanceID))
	}
	req := client.uhostconn.NewStopUHostInstanceRequest()
	req.UHostId = ucloud.String(instanceID)
	_, err := client.uhostconn.StopUHostInstance(req)
	return err
}

func (client *productClient) startInstanceByID(instanceID string) error {
	if instanceID == "" {
		return newNotFoundError(getNotFoundMessage("instance", instanceID))
	}
	req := client.uhostconn.NewStartUHostInstanceRequest()
	req.UHostId = ucloud.String(instanceID)
	_, err := client.uhostconn.StartUHostInstance(req)
	return err
}
