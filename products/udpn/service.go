package udpn

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func describeDPNById(client *udpn.UDPNClient, id string) (*udpn.UDPNData, error) {
	if id == "" {
		return nil, newNotFoundError(getNotFoundMessage("dpn", id))
	}

	req := client.NewDescribeUDPNRequest()
	req.UDPNId = ucloud.String(id)

	resp, err := client.DescribeUDPN(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading dpn %q, %s", id, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("dpn", id))
	}

	return &resp.DataSet[0], nil
}
