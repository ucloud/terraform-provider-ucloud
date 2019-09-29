package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (c *UCloudClient) describeVPNGatewayById(vpnGwId string) (*ipsecvpn.VPNGatewayDataSet, error) {
	if vpnGwId == "" {
		return nil, newNotFoundError(getNotFoundMessage("vpn_gateway", vpnGwId))
	}
	conn := c.ipsecvpnClient

	req := conn.NewDescribeVPNGatewayRequest()
	req.VPNGatewayIds = []string{vpnGwId}

	resp, err := conn.DescribeVPNGateway(req)

	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("vpn_gateway", vpnGwId))
		}
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("vpn_gateway", vpnGwId))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeVPNCustomerGatewayById(vpnCusGwId string) (*ipsecvpn.RemoteVPNGatewayDataSet, error) {
	if vpnCusGwId == "" {
		return nil, newNotFoundError(getNotFoundMessage("vpn_costomer_gateway", vpnCusGwId))
	}
	conn := c.ipsecvpnClient

	req := conn.NewDescribeRemoteVPNGatewayRequest()
	req.RemoteVPNGatewayIds = []string{vpnCusGwId}

	resp, err := conn.DescribeRemoteVPNGateway(req)

	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("vpn_costomer_gateway", vpnCusGwId))
		}
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("vpn_costomer_gateway", vpnCusGwId))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeVPNConnectionById(vpnConnId string) (*ipsecvpn.VPNTunnelDataSet, error) {
	if vpnConnId == "" {
		return nil, newNotFoundError(getNotFoundMessage("vpn_connection", vpnConnId))
	}
	conn := c.ipsecvpnClient

	req := conn.NewDescribeVPNTunnelRequest()
	req.VPNTunnelIds = []string{vpnConnId}

	resp, err := conn.DescribeVPNTunnel(req)

	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 54002 {
			return nil, newNotFoundError(getNotFoundMessage("vpn_connection", vpnConnId))
		}
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("vpn_connection", vpnConnId))
	}

	return &resp.DataSet[0], nil
}
