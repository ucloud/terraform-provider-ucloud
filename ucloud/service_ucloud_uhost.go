package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
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

func (client *UCloudClient) DescribeImageById(imageId string) (*uhost.UHostImageSet, error) {
	req := client.uhostconn.NewDescribeImageRequest()
	req.ImageId = ucloud.String(imageId)

	resp, err := client.uhostconn.DescribeImage(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ImageSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("instance", imageId))
	}

	return &resp.ImageSet[0], nil
}
