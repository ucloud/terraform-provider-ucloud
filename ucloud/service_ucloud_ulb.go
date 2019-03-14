package ucloud

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *UCloudClient) describeLBById(lbId string) (*ulb.ULBSet, error) {
	conn := client.ulbconn
	req := conn.NewDescribeULBRequest()
	req.ULBId = ucloud.String(lbId)

	resp, err := conn.DescribeULB(req)

	// [API-STYLE] lb api has not found err code, but others don't have
	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && (uErr.Code() == 4103 || uErr.Code() == 4086) {
			return nil, newNotFoundError(getNotFoundMessage("lb", lbId))
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("lb", lbId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) describeVServerById(lbId, listenerId string) (*ulb.ULBVServerSet, error) {
	conn := client.ulbconn
	req := conn.NewDescribeVServerRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(listenerId)

	resp, err := conn.DescribeVServer(req)

	// [API-STYLE] vserver api has not found err code, but others don't have
	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 4103 {
			return nil, newNotFoundError(getNotFoundMessage("listener", listenerId))
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("listener", listenerId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) describeBackendById(lbId, listenerId, backendId string) (*ulb.ULBBackendSet, error) {
	vserverSet, err := client.describeVServerById(lbId, listenerId)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(vserverSet.BackendSet); i++ {
		backend := vserverSet.BackendSet[i]
		if backend.BackendId == backendId {
			return &backend, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("backend", backendId))
}

func (client *UCloudClient) describePolicyById(lbId, listenerId, policyId string) (*ulb.ULBPolicySet, error) {
	vserverSet, err := client.describeVServerById(lbId, listenerId)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(vserverSet.PolicySet); i++ {
		policy := vserverSet.PolicySet[i]
		if policy.PolicyId == policyId {
			return &policy, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("policy", policyId))
}

func (client *UCloudClient) describeLBSSLById(sslId string) (*ulb.ULBSSLSet, error) {
	conn := client.ulbconn
	req := conn.NewDescribeSSLRequest()
	req.SSLId = ucloud.String(sslId)

	resp, err := conn.DescribeSSL(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("lb_ssl", sslId))
	}

	return &resp.DataSet[0], nil
}

func (c *UCloudClient) describeLBSSLAttachmentById(sslId, ulbId, vserverId string) (*ulb.SSLBindedTargetSet, error) {
	conn := c.ulbconn

	req := conn.NewDescribeSSLRequest()
	req.SSLId = ucloud.String(sslId)

	resp, err := conn.DescribeSSL(req)
	if err != nil {
		return nil, err
	}

	if resp == nil || len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("lb_ssl_attachment", sslId))
	}

	for i := 0; i < len(resp.DataSet); i++ {
		ssl := resp.DataSet[i]
		for m := 0; m < len(ssl.BindedTargetSet); m++ {
			if ssl.BindedTargetSet[m].ULBId == ulbId && ssl.BindedTargetSet[m].VServerId == vserverId {
				return &ssl.BindedTargetSet[m], nil
			}
		}

	}

	return nil, newNotFoundError(getNotFoundMessage("lb_ssl_attachment", sslId))
}

func (client *UCloudClient) describeVServerByOneId(listenerId string) (*ulb.ULBVServerSet, string, error) {
	conn := client.ulbconn
	req := conn.NewDescribeVServerRequest()

	lbId, err := client.getLBIdFromVServerId(listenerId)
	if err != nil {
		return nil, "", newNotFoundError(getNotFoundMessage("listener", listenerId))
	}
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(listenerId)

	resp, err := conn.DescribeVServer(req)

	// [API-STYLE] vserver api has not found err code, but others don't have
	// TODO: don't use magic number
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 4103 {
			return nil, "", newNotFoundError(getNotFoundMessage("listener", listenerId))
		}
		return nil, "", err
	}

	if len(resp.DataSet) < 1 {
		return nil, "", newNotFoundError(getNotFoundMessage("listener", listenerId))
	}

	return &resp.DataSet[0], lbId, nil
}

func (client *UCloudClient) getLBIdFromVServerId(listenerId string) (string, error) {
	conn := client.ulbconn
	req := conn.NewDescribeULBRequest()

	var ulbSets []ulb.ULBSet
	var limit int = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeULB(req)
		if err != nil {
			return "", err
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		ulbSets = append(ulbSets, resp.DataSet...)

		for _, ulbSet := range ulbSets {
			for _, v := range ulbSet.VServerSet {
				if v.VServerId == listenerId {
					return ulbSet.ULBId, nil
				}
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	return "", fmt.Errorf("parse failed")
}

func (client *UCloudClient) describeBackendByOneId(backendId string) (*ulb.ULBBackendSet, string, string, error) {
	var err error
	lbId, listenerId, err := client.getLBIdAndVServerIdFromBackendId(backendId)
	if err != nil {
		return nil, "", "", newNotFoundError(getNotFoundMessage("listener", backendId))
	}

	vserverSet, err := client.describeVServerById(lbId, listenerId)
	if err != nil {
		return nil, "", "", err
	}

	for i := 0; i < len(vserverSet.BackendSet); i++ {
		backend := vserverSet.BackendSet[i]
		if backend.BackendId == backendId {
			return &backend, lbId, listenerId, nil
		}
	}

	return nil, "", "", newNotFoundError(getNotFoundMessage("backend", backendId))
}

func (client *UCloudClient) getLBIdAndVServerIdFromBackendId(backendId string) (string, string, error) {
	conn := client.ulbconn
	req := conn.NewDescribeULBRequest()

	var ulbSets []ulb.ULBSet
	var limit int = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeULB(req)
		if err != nil {
			return "", "", err
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		ulbSets = append(ulbSets, resp.DataSet...)

		for _, ulbSet := range ulbSets {
			for _, vserverSet := range ulbSet.VServerSet {
				for _, v := range vserverSet.BackendSet {
					if v.BackendId == backendId {
						return ulbSet.ULBId, vserverSet.VServerId, nil
					}
				}
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	return "", "", fmt.Errorf("parse failed")
}

func (client *UCloudClient) describePolicyByOneId(policyId string) (*ulb.ULBPolicySet, string, string, error) {
	var err error
	lbId, listenerId, err := client.getLBIdAndVServerIdFromPolicyId(policyId)
	if err != nil {
		return nil, "", "", newNotFoundError(getNotFoundMessage("policy", policyId))
	}

	vserverSet, err := client.describeVServerById(lbId, listenerId)
	if err != nil {
		return nil, "", "", err
	}

	for i := 0; i < len(vserverSet.PolicySet); i++ {
		policy := vserverSet.PolicySet[i]
		if policy.PolicyId == policyId {
			return &policy, lbId, listenerId, nil
		}
	}

	return nil, "", "", newNotFoundError(getNotFoundMessage("policy", policyId))
}

func (client *UCloudClient) getLBIdAndVServerIdFromPolicyId(policyId string) (string, string, error) {
	conn := client.ulbconn
	req := conn.NewDescribeULBRequest()

	var ulbSets []ulb.ULBSet
	var limit int = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeULB(req)
		if err != nil {
			return "", "", err
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		ulbSets = append(ulbSets, resp.DataSet...)

		for _, ulbSet := range ulbSets {
			for _, vserverSet := range ulbSet.VServerSet {
				for _, v := range vserverSet.PolicySet {
					if v.PolicyId == policyId {
						return ulbSet.ULBId, vserverSet.VServerId, nil
					}
				}
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	return "", "", fmt.Errorf("parse failed")
}
