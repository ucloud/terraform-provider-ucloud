package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *UCloudClient) describeDPNById(id string) (*udpn.UDPNData, error) {
	conn := c.udpnconn

	req := conn.NewDescribeUDPNRequest()
	req.UDPNId = ucloud.String(id)

	resp, err := conn.DescribeUDPN(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("dpn", id))
	}

	return &resp.DataSet[0], nil
}
