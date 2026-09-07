package vpc

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (c *productClient) describeSecGroupsByVPCId(vpcID string) ([]vpc.SecGroupInfo, error) {
	conn := c.vpcconn

	var allSecGroups []vpc.SecGroupInfo
	var offset int
	const limit = 100

	for {
		req := conn.NewDescribeSecGroupRequest()
		if vpcID != "" {
			req.VPCId = ucloud.String(vpcID)
		}
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeSecGroup(req)
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.GetRetCode() != 0 {
			return nil, fmt.Errorf("error on reading sec group list, %s", resp.GetMessage())
		}
		if resp == nil {
			break
		}

		allSecGroups = append(allSecGroups, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
		offset += limit
	}

	return allSecGroups, nil
}

func (c *productClient) describeResourceSecGroup(resourceID string) ([]vpc.BindingSecGroupInfo, error) {
	conn := c.vpcconn

	req := conn.NewDescribeResourceSecGroupRequest()
	req.ResourceId = []string{resourceID}

	resp, err := conn.DescribeResourceSecGroup(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading sec group bindings for resource %q, %s", resourceID, resp.GetMessage())
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, nil
	}

	return resp.DataSet[0].SecGroupInfo, nil
}
