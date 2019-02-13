package ucloud

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (c *UCloudClient) describeEIPById(eipId string) (*unet.UnetEIPSet, error) {
	conn := c.unetconn

	req := conn.NewDescribeEIPRequest()
	req.EIPIds = []string{eipId}

	resp, err := conn.DescribeEIP(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.EIPSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("eip", eipId))
	}

	return &resp.EIPSet[0], nil
}

func (c *UCloudClient) describeEIPResourceById(eipId, resourceId string) (*unet.UnetEIPResourceSet, error) {
	conn := c.unetconn

	req := conn.NewDescribeEIPRequest()
	req.EIPIds = []string{eipId}

	resp, err := conn.DescribeEIP(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.EIPSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("eip association", eipId))
	}

	for i := 0; i < len(resp.EIPSet); i++ {
		eip := resp.EIPSet[i]
		if eip.Resource.ResourceId == resourceId {
			return &eip.Resource, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("eip association", eipId))
}

func (c *UCloudClient) checkDefaultFirewall() error {
	conn := c.unetconn

	req := conn.NewDescribeFirewallRequest()

	resp, err := conn.DescribeFirewall(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.DataSet) < 2 {
		return fmt.Errorf("the default firewall is not found for this project/region, it will be initializing automiticly, please try again later")
	}

	return nil
}

func (c *UCloudClient) describeFirewallById(sgId string) (*unet.FirewallDataSet, error) {
	conn := c.unetconn

	req := conn.NewDescribeFirewallRequest()
	req.FWId = ucloud.String(sgId)

	resp, err := conn.DescribeFirewall(req)

	// [API-STYLE] Fire wall api has not found err code, but others don't have
	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("security group", sgId))
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("security group", sgId))
	}

	return &resp.DataSet[0], nil
}
