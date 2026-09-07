package unet

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func describeEIPById(c *sdkClient, eipId string) (*unet.UnetEIPSet, error) {
	if eipId == "" {
		return nil, newNotFoundError(getNotFoundMessage("eip", eipId))
	}
	conn := c

	req := conn.NewDescribeEIPRequest()
	req.EIPIds = []string{eipId}

	resp, err := conn.DescribeEIP(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading eip %q, %s", eipId, resp.GetMessage())
	}
	if resp == nil || len(resp.EIPSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("eip", eipId))
	}

	return &resp.EIPSet[0], nil
}

func describeEIPResourceById(c *sdkClient, eipId, resourceId string) (*unet.UnetEIPResourceSet, error) {
	conn := c

	req := conn.NewDescribeEIPRequest()
	req.EIPIds = []string{eipId}

	resp, err := conn.DescribeEIP(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading eip association %q, %s", eipId, resp.GetMessage())
	}
	if resp == nil || len(resp.EIPSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("eip association", eipId))
	}

	for i := 0; i < len(resp.EIPSet); i++ {
		eip := resp.EIPSet[i]
		if eip.Resource.ResourceID == resourceId {
			return &eip.Resource, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("eip association", eipId))
}

func describeFirewallById(c *sdkClient, sgId string) (*unet.FirewallDataSet, error) {
	if sgId == "" {
		return nil, newNotFoundError(getNotFoundMessage("security group", sgId))
	}
	conn := c

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
