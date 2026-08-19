package ucloud

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

// DescribeSecGroup answers a missing sec group with this error code instead of
// an empty result set, eg. "describe secgroup error: secgroup-xxx not exist in VPC()"
const secGroupNotExistCode = 208704

func (c *UCloudClient) describeSecGroupsByVPCId(vpcId string) ([]vpc.SecGroupInfo, error) {
	conn := c.vpcconn

	var allSecGroups []vpc.SecGroupInfo
	var offset int
	var limit = 100

	for {
		req := conn.NewDescribeSecGroupRequest()
		if vpcId != "" {
			req.VPCId = ucloud.String(vpcId)
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

func (c *UCloudClient) describeResourceSecGroup(resourceId string) ([]vpc.BindingSecGroupInfo, error) {
	conn := c.vpcconn

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

func (c *UCloudClient) describeSecGroupById(secGroupId string) (*vpc.SecGroupInfo, error) {
	if secGroupId == "" {
		return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupId))
	}

	conn := c.vpcconn

	req := conn.NewDescribeSecGroupRequest()
	req.SecGroupId = []string{secGroupId}

	resp, err := conn.DescribeSecGroup(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == secGroupNotExistCode {
			return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupId))
		}
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		if resp.GetRetCode() == secGroupNotExistCode {
			return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupId))
		}
		return nil, fmt.Errorf("error on reading sec group %q, %s", secGroupId, resp.GetMessage())
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupId))
	}

	return &resp.DataSet[0], nil
}

// there is no api to describe a single sec group rule, so we read the whole
// sec group and pick the rule out of it by rule id
func (c *UCloudClient) describeSecGroupRuleById(secGroupId, ruleId string) (*vpc.SecGroupRuleInfo, error) {
	if secGroupId == "" || ruleId == "" {
		return nil, newNotFoundError(getNotFoundMessage("sec group rule", ruleId))
	}

	sgSet, err := c.describeSecGroupById(secGroupId)
	if err != nil {
		return nil, err
	}

	for i := range sgSet.Rule {
		if sgSet.Rule[i].RuleId == ruleId {
			return &sgSet.Rule[i], nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("sec group rule", ruleId))
}
