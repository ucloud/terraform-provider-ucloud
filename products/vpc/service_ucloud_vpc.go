package vpc

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (c *productClient) describeVPCById(vpcID string) (*vpc.VPCInfo, error) {
	if vpcID == "" {
		return nil, newNotFoundError(getNotFoundMessage("vpc", vpcID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeVPCRequest()
	req.VPCIds = []string{vpcID}

	resp, err := conn.DescribeVPC(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading vpc %q, %s", vpcID, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("vpc", vpcID))
	}

	return &resp.DataSet[0], nil
}

func (c *productClient) describeSubnetById(subnetID string) (*vpc.SubnetInfo, error) {
	if subnetID == "" {
		return nil, newNotFoundError(getNotFoundMessage("subnet", subnetID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeSubnetRequest()
	req.SubnetIds = []string{subnetID}

	resp, err := conn.DescribeSubnet(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading subnet %q, %s", subnetID, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("subnet", subnetID))
	}

	return &resp.DataSet[0], nil
}

func (c *productClient) describeVPCIntercomById(vpcID, peerVPCID, peerRegion, peerProjectID string) (*vpc.VPCIntercomInfo, error) {
	conn := c.vpcconn

	req := conn.NewDescribeVPCIntercomRequest()
	req.VPCId = ucloud.String(vpcID)
	req.DstRegion = ucloud.String(peerRegion)
	req.DstProjectId = ucloud.String(peerProjectID)

	resp, err := conn.DescribeVPCIntercom(req)
	if err != nil {
		if uCloudErr, ok := err.(uerr.Error); ok && uCloudErr.Code() == 58103 {
			return nil, newNotFoundError(getNotFoundMessage("vpc peer connection", vpcID))
		}
		return nil, err
	}

	for i := 0; i < len(resp.DataSet); i++ {
		if resp.DataSet[i].VPCId == peerVPCID {
			return &resp.DataSet[0], nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("vpc peer connection", vpcID))
}

func (c *productClient) describeNatGatewayById(natGatewayID string) (*vpc.NatGatewayDataSet, error) {
	if natGatewayID == "" {
		return nil, newNotFoundError(getNotFoundMessage("nat_gateway", natGatewayID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeNATGWRequest()
	req.NATGWIds = []string{natGatewayID}

	resp, err := conn.DescribeNATGW(req)
	if err != nil {
		if uCloudErr, ok := err.(uerr.Error); ok && uCloudErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("nat_gateway", natGatewayID))
		}
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("nat_gateway", natGatewayID))
	}

	return &resp.DataSet[0], nil
}

func (c *productClient) describeNatGatewayRuleById(policyID, natGatewayID string) (*vpc.NATGWPolicyDataSet, error) {
	if policyID == "" {
		return nil, newNotFoundError(getNotFoundMessage("nat_gateway_rule", policyID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeNATGWPolicyRequest()
	req.NATGWId = ucloud.String(natGatewayID)

	resp, err := conn.DescribeNATGWPolicy(req)
	if err != nil {
		if uCloudErr, ok := err.(uerr.Error); ok && uCloudErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("nat_gateway_rule", policyID))
		}
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("nat_gateway_rule", policyID))
	}

	for i := 0; i < len(resp.DataSet); i++ {
		policySet := resp.DataSet[i]
		if policySet.PolicyId == policyID {
			return &policySet, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("nat_gateway_rule", policyID))
}

func (c *productClient) describeVIPById(vipID string) (*vpc.VIPDetailSet, error) {
	if vipID == "" {
		return nil, newNotFoundError(getNotFoundMessage("vip", vipID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeVIPRequest()
	req.VIPId = ucloud.String(vipID)

	resp, err := conn.DescribeVIP(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading vip %q, %s", vipID, resp.GetMessage())
	}
	if resp == nil || len(resp.VIPSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("vip", vipID))
	}

	return &resp.VIPSet[0], nil
}
