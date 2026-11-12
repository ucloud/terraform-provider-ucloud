package vpc

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

// describeSecGroupById reads a single security group and reports a missing one
// as a not-found error. describeSecGroupsByVPCId cannot be used for this: it
// serves the data source and treats an empty result as an empty list.
//
// No pagination is needed: DescribeSecGroup documents SecGroupId as taking over
// from Offset and Limit when it is passed, so the group comes back whole, rules
// included. That is what lets a rule be looked up by scanning this result.
func (c *productClient) describeSecGroupById(secGroupID string) (*vpc.SecGroupInfo, error) {
	if secGroupID == "" {
		return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupID))
	}
	conn := c.vpcconn

	req := conn.NewDescribeSecGroupRequest()
	req.SecGroupId = []string{secGroupID}

	resp, err := conn.DescribeSecGroup(req)
	if err != nil {
		// A group that does not exist is reported as a server error, not as an
		// empty DataSet: the SDK turns any non-zero retcode into a uerr.Error
		// before it reaches us (see the client's errorHandler). That error has
		// to be translated here, because isNotFoundError only knows the
		// provider's own not-found type, and every caller leans on it to tell
		// "the group is gone" apart from "the call failed" - a read that
		// cannot tell them apart fails instead of clearing the ID, and a
		// delete that cannot tell them apart stops being idempotent.
		//
		// Two codes mean gone here. 54002 is the generic one the rest of the
		// provider reads; the security group backend reports a missing group
		// with 208704 of its own, which is what a delete sees on the retry that
		// follows its own delete call. Both are accepted so the check holds
		// whichever layer answers, and 54002 is kept rather than replaced
		// because dropping it would bet on the backend never using it.
		if uCloudErr, ok := err.(uerr.Error); ok &&
			(uCloudErr.Code() == 54002 || uCloudErr.Code() == 208704) {
			return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupID))
		}
		return nil, err
	}
	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("sec group", secGroupID))
	}

	return &resp.DataSet[0], nil
}

// describeSecGroupRuleById reads a single rule of a security group and reports a
// missing one as a not-found error. DescribeSecGroup has no RuleId filter, so
// the parent group is fetched and its rules searched linearly; a rule cannot be
// located from its own ID alone.
func (c *productClient) describeSecGroupRuleById(secGroupID, ruleID string) (*vpc.SecGroupRuleInfo, error) {
	if secGroupID == "" || ruleID == "" {
		return nil, newNotFoundError(getNotFoundMessage("sec group rule", ruleID))
	}

	secGroup, err := c.describeSecGroupById(secGroupID)
	if err != nil {
		return nil, err
	}
	for i := range secGroup.Rule {
		if secGroup.Rule[i].RuleId == ruleID {
			return &secGroup.Rule[i], nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("sec group rule", ruleID))
}

func (c *productClient) describeSecGroupsByVPCId(vpcID string) ([]vpc.SecGroupInfo, error) {
	conn := c.vpcconn

	var offset int
	const limit = 100
	allSecGroups := make([]vpc.SecGroupInfo, 0, limit)

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
