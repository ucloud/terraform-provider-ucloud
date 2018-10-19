package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (c *UCloudClient) describeVPCById(vpcId string) (*vpc.VPCInfo, error) {
	conn := c.vpcconn

	req := conn.NewDescribeVPCRequest()
	req.VPCIds = []string{vpcId}

	resp, err := conn.DescribeVPC(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("vpc", vpcId))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeSubnetById(subnetId string) (*vpc.VPCSubnetInfoSet, error) {
	conn := c.vpcconn

	req := conn.NewDescribeSubnetRequest()
	req.SubnetIds = []string{subnetId}

	resp, err := conn.DescribeSubnet(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("subnet", subnetId))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeVPCIntercomById(vpcId, peerVPCId, peerRegion, peerProjectId string) (*vpc.VPCIntercomInfo, error) {
	conn := c.vpcconn

	req := conn.NewDescribeVPCIntercomRequest()
	req.VPCId = ucloud.String(vpcId)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectId)

	resp, err := conn.DescribeVPCIntercom(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 58103 {
			return nil, newNotFoundError(getNotFoundMessage("vpc peer connection", vpcId))
		}
		return nil, err
	}

	for i := 0; i < len(resp.DataSet); i++ {
		if resp.DataSet[i].VPCId == peerVPCId {
			return &resp.DataSet[0], nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("vpc peer connection", vpcId))
}
