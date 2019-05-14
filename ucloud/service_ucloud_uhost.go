package ucloud

import (
	"strconv"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (client *UCloudClient) describeInstanceById(instanceId string) (*uhost.UHostInstanceSet, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
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
	if imageId == "" {
		return nil, newNotFoundError(getNotFoundMessage("image", imageId))
	}
	req := client.uhostconn.NewDescribeImageRequest()
	req.ImageId = ucloud.String(imageId)

	resp, err := client.uhostconn.DescribeImage(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ImageSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("image", imageId))
	}

	return &resp.ImageSet[0], nil
}

func instanceTypeSetFunc(machineType string, cpu, memory int) string {
	if memory/cpu == 1 {
		return strings.Join([]string{machineType, "highcpu", strconv.Itoa(cpu)}, "-")
	}

	if memory/cpu == 2 {
		return strings.Join([]string{machineType, "basic", strconv.Itoa(cpu)}, "-")
	}

	if memory/cpu == 4 {
		return strings.Join([]string{machineType, "standard", strconv.Itoa(cpu)}, "-")
	}

	if memory/cpu == 8 {
		return strings.Join([]string{machineType, "highmem", strconv.Itoa(cpu)}, "-")
	}

	return strings.Join([]string{"n", "customized", strconv.Itoa(cpu), strconv.Itoa(memory)}, "-")
}
