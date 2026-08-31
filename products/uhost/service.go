package uhost

import (
	"fmt"
	"strconv"
	"strings"

	sdkuhost "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *productClient) describeInstanceById(instanceId string) (*sdkuhost.UHostInstanceSet, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	req := client.uhostconn.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{instanceId}

	resp, err := client.uhostconn.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading instance %q, %s", instanceId, resp.GetMessage())
	}
	if len(resp.UHostSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
	}

	return &resp.UHostSet[0], nil
}

func (client *productClient) describeImageById(imageId string) (*sdkuhost.UHostImageSet, error) {
	if imageId == "" {
		return nil, newNotFoundError(getNotFoundMessage("image", imageId))
	}
	req := client.uhostconn.NewDescribeImageRequest()
	req.ImageId = ucloud.String(imageId)

	resp, err := client.uhostconn.DescribeImage(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading image %q, %s", imageId, resp.GetMessage())
	}
	if len(resp.ImageSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("image", imageId))
	}

	return &resp.ImageSet[0], nil
}

func (client *productClient) describeIsolationGroupById(igId string) (*sdkuhost.IsolationGroup, error) {
	if igId == "" {
		return nil, newNotFoundError(getNotFoundMessage("isolation group", igId))
	}
	req := client.uhostconn.NewDescribeIsolationGroupRequest()
	req.GroupId = ucloud.String(igId)

	resp, err := client.uhostconn.DescribeIsolationGroup(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 8037 {
			return nil, newNotFoundError(getNotFoundMessage("isolation group", igId))
		}
		return nil, err
	}

	if len(resp.IsolationGroupSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("isolation group", igId))
	}

	return &resp.IsolationGroupSet[0], nil
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

func (client *productClient) describeFirewallByIdAndType(resourceId, resourceType string) (*unet.FirewallDataSet, error) {
	conn := client.unetconn

	req := conn.NewDescribeFirewallRequest()
	req.ResourceId = ucloud.String(resourceId)
	req.ResourceType = ucloud.String(resourceType)

	resp, err := conn.DescribeFirewall(req)
	// [API-STYLE] Fire wall api has not found err code, but others don't have
	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, newNotFoundError("")
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError("")
	}

	return &resp.DataSet[0], nil
}

func (client *productClient) getInstanceState(instanceId string) (string, error) {
	if instanceId == "" {
		return "", newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	instance, err := client.describeInstanceById(instanceId)
	if err != nil {
		return "", fmt.Errorf("fail to get instance info: %w", err)
	}
	return instance.State, nil
}

func (client *productClient) startInstanceById(instanceId string) error {
	if instanceId == "" {
		return newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	req := client.uhostconn.NewStartUHostInstanceRequest()
	req.UHostId = ucloud.String(instanceId)
	_, err := client.uhostconn.StartUHostInstance(req)
	return err
}

func (client *productClient) stopInstanceById(instanceId string) error {
	if instanceId == "" {
		return newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	req := client.uhostconn.NewStopUHostInstanceRequest()
	req.UHostId = ucloud.String(instanceId)
	_, err := client.uhostconn.StopUHostInstance(req)
	return err
}

func (client *productClient) poweroffInstanceById(instanceId string) error {
	if instanceId == "" {
		return newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	req := client.uhostconn.NewPoweroffUHostInstanceRequest()
	req.UHostId = ucloud.String(instanceId)
	_, err := client.uhostconn.PoweroffUHostInstance(req)
	return err
}

func (client *productClient) describeResourceSecGroup(resourceId string) ([]vpc.BindingSecGroupInfo, error) {
	conn := client.vpcconn

	req := conn.NewDescribeResourceSecGroupRequest()
	req.ResourceId = []string{resourceId}

	resp, err := conn.DescribeResourceSecGroup(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading sec group bindings for resource %q, %s", resourceId, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, nil
	}

	return resp.DataSet[0].SecGroupInfo, nil
}

func (client *productClient) checkDefaultFirewall() error {
	conn := client.unetconn

	req := conn.NewDescribeFirewallRequest()

	resp, err := conn.DescribeFirewall(req)
	if err != nil {
		return fmt.Errorf("error on reading default security group before creating instance, %s", err)
	}

	if resp == nil || len(resp.DataSet) < 2 {
		return fmt.Errorf("the default security group is not found for this project/region, it will be initializing automiticly, please try again later")
	}

	return nil
}
