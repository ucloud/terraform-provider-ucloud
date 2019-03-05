package ucloud

import (
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
