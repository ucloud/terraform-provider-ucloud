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

func (c *productClient) describeRouteTableById(routeTableID string) (*vpc.RouteTableInfo, error) {
	if routeTableID == "" {
		return nil, newNotFoundError(getNotFoundMessage("route_table", routeTableID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeRouteTableRequest()
	req.RouteTableId = ucloud.String(routeTableID)

	resp, err := conn.DescribeRouteTable(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading route table %q, %s", routeTableID, resp.GetMessage())
	}
	if resp == nil || len(resp.RouteTables) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("route_table", routeTableID))
	}

	return &resp.RouteTables[0], nil
}

func (c *productClient) describeRouteRuleById(routeTableID, routeRuleID string) (*vpc.RouteRuleInfo, error) {
	if routeTableID == "" || routeRuleID == "" {
		return nil, newNotFoundError(getNotFoundMessage("route_table_rule", routeRuleID))
	}
	routeTable, err := c.describeRouteTableById(routeTableID)
	if err != nil {
		return nil, err
	}
	for i := range routeTable.RouteRules {
		if routeTable.RouteRules[i].RouteRuleId == routeRuleID {
			return &routeTable.RouteRules[i], nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("route_table_rule", routeRuleID))
}

func (c *productClient) describeNetworkInterfaceById(interfaceID string) (*vpc.NetworkInterface, error) {
	if interfaceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("network_interface", interfaceID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeNetworkInterfaceRequest()
	req.InterfaceId = []string{interfaceID}

	resp, err := conn.DescribeNetworkInterface(req)
	if err != nil {
		if uCloudErr, ok := err.(uerr.Error); ok && uCloudErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("network_interface", interfaceID))
		}
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading network interface %q, %s", interfaceID, resp.GetMessage())
	}
	if resp == nil || len(resp.NetworkInterfaceSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("network_interface", interfaceID))
	}

	return &resp.NetworkInterfaceSet[0], nil
}

// modifyNetworkInterfaceAttribute updates the name/tag/remark extend info of a
// network interface via the ModifyNetworkInterface action. The vendored SDK
// (v0.22.70) does not expose this action as a typed client method, so it is
// invoked through the SDK generic client. The backend only updates non-empty
// fields, so empty values are ignored rather than cleared.
func (c *productClient) modifyNetworkInterfaceAttribute(interfaceID, name, tag, remark string) error {
	req := c.vpcconn.NewGenericRequest()
	if err := req.SetPayload(map[string]interface{}{
		"Action":      "ModifyNetworkInterface",
		"InterfaceId": interfaceID,
		"Name":        name,
		"Tag":         tag,
		"Remark":      remark,
	}); err != nil {
		return err
	}
	resp, err := c.vpcconn.GenericInvoke(req)
	if err != nil {
		return err
	}
	if resp.GetRetCode() != 0 {
		return fmt.Errorf("error on modifying network interface %q, %s", interfaceID, resp.GetMessage())
	}
	return nil
}

func (c *productClient) describeRouteRuleByAttrs(routeTableID, dstAddr, nexthopID string) (*vpc.RouteRuleInfo, error) {
	routeTable, err := c.describeRouteTableById(routeTableID)
	if err != nil {
		return nil, err
	}
	for i := range routeTable.RouteRules {
		rule := routeTable.RouteRules[i]
		if rule.DstAddr == dstAddr && rule.NexthopId == nexthopID {
			return &rule, nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("route_table_rule", dstAddr+" "+nexthopID))
}
