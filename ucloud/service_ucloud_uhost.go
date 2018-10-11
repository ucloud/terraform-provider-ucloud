package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

func (client *UCloudClient) describeInstanceById(instanceId string) (*uhost.UHostInstanceSet, error) {
	req := client.uhostconn.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{instanceId}

	resp, err := client.uhostconn.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
	}

	return &resp.UHostSet[0], nil
}
