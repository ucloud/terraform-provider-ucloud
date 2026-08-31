package ulb

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *productClient) describeLBById(lbID string) (*ulb.ULBSet, error) {
	if lbID == "" {
		return nil, newNotFoundError(getNotFoundMessage("lb", lbID))
	}
	request := client.ulbconn.NewDescribeULBRequest()
	request.ULBId = ucloud.String(lbID)
	response, err := client.ulbconn.DescribeULB(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && (cloudErr.Code() == 4103 || cloudErr.Code() == 4086) {
			return nil, newNotFoundError(getNotFoundMessage("lb", lbID))
		}
		return nil, err
	}
	if response == nil || len(response.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("lb", lbID))
	}
	return &response.DataSet[0], nil
}

func (client *productClient) describeVServerById(lbID, listenerID string) (*ulb.ULBVServerSet, error) {
	if listenerID == "" {
		return nil, newNotFoundError(getNotFoundMessage("listener", listenerID))
	}
	request := client.ulbconn.NewDescribeVServerRequest()
	request.ULBId = ucloud.String(lbID)
	request.VServerId = ucloud.String(listenerID)
	response, err := client.ulbconn.DescribeVServer(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 4103 {
			return nil, newNotFoundError(getNotFoundMessage("listener", listenerID))
		}
		return nil, err
	}
	if response == nil || len(response.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("listener", listenerID))
	}
	return &response.DataSet[0], nil
}

func (client *productClient) describeBackendById(lbID, listenerID, backendID string) (*ulb.ULBBackendSet, error) {
	if backendID == "" {
		return nil, newNotFoundError(getNotFoundMessage("lb_attachment", backendID))
	}
	vserverSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return nil, err
	}
	for i := range vserverSet.BackendSet {
		backend := vserverSet.BackendSet[i]
		if backend.BackendId == backendID {
			return &backend, nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("lb_attachment", backendID))
}

func (client *productClient) describePolicyById(lbID, listenerID, policyID string) (*ulb.ULBPolicySet, error) {
	if policyID == "" {
		return nil, newNotFoundError(getNotFoundMessage("policy", policyID))
	}
	vserverSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return nil, err
	}
	for i := range vserverSet.PolicySet {
		policy := vserverSet.PolicySet[i]
		if policy.PolicyId == policyID {
			return &policy, nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("policy", policyID))
}

func (client *productClient) describeLBSSLById(sslID string) (*ulb.ULBSSLSet, error) {
	if sslID == "" {
		return nil, newNotFoundError(getNotFoundMessage("lb_ssl", sslID))
	}
	request := client.ulbconn.NewDescribeSSLRequest()
	request.SSLId = ucloud.String(sslID)
	response, err := client.ulbconn.DescribeSSL(request)
	if err != nil {
		return nil, err
	}
	if response == nil || response.GetRetCode() != 0 {
		if response == nil {
			return nil, newNotFoundError(getNotFoundMessage("lb_ssl", sslID))
		}
		return nil, fmt.Errorf("error on reading lb_ssl %q, %s", sslID, response.GetMessage())
	}
	if len(response.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("lb_ssl", sslID))
	}
	return &response.DataSet[0], nil
}

func (client *productClient) describeLBSSLAttachmentById(sslID, lbID, vserverID string) (*ulb.SSLBindedTargetSet, error) {
	request := client.ulbconn.NewDescribeSSLRequest()
	request.SSLId = ucloud.String(sslID)
	response, err := client.ulbconn.DescribeSSL(request)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, newNotFoundError(getNotFoundMessage("lb_ssl_attachment", sslID))
	}
	if response.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading lb_ssl_attachment %q, %s", sslID, response.GetMessage())
	}
	if len(response.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("lb_ssl_attachment", sslID))
	}
	for i := range response.DataSet {
		ssl := response.DataSet[i]
		for m := range ssl.BindedTargetSet {
			if ssl.BindedTargetSet[m].ULBId == lbID && ssl.BindedTargetSet[m].VServerId == vserverID {
				return &ssl.BindedTargetSet[m], nil
			}
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("lb_ssl_attachment", sslID))
}

func (client *productClient) describeVServerByOneId(listenerID string) (*ulb.ULBVServerSet, string, error) {
	request := client.ulbconn.NewDescribeVServerRequest()
	lbID, err := client.getLBIdFromVServerId(listenerID)
	if err != nil {
		return nil, "", err
	}
	request.ULBId = ucloud.String(lbID)
	request.VServerId = ucloud.String(listenerID)
	response, err := client.ulbconn.DescribeVServer(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 4103 {
			return nil, "", newNotFoundError(getNotFoundMessage("listener", listenerID))
		}
		return nil, "", err
	}
	if response == nil || len(response.DataSet) < 1 {
		return nil, "", newNotFoundError(getNotFoundMessage("listener", listenerID))
	}
	return &response.DataSet[0], lbID, nil
}

func (client *productClient) getLBIdFromVServerId(listenerID string) (string, error) {
	request := client.ulbconn.NewDescribeULBRequest()
	var ulbSets []ulb.ULBSet
	const limit = 100
	for offset := 0; ; offset += limit {
		request.Limit = ucloud.Int(limit)
		request.Offset = ucloud.Int(offset)
		response, err := client.ulbconn.DescribeULB(request)
		if err != nil {
			return "", err
		}
		if response == nil || len(response.DataSet) < 1 {
			break
		}
		ulbSets = append(ulbSets, response.DataSet...)
		for _, lb := range ulbSets {
			for _, listener := range lb.VServerSet {
				if listener.VServerId == listenerID {
					return lb.ULBId, nil
				}
			}
		}
		if len(response.DataSet) < limit {
			break
		}
	}
	return "", fmt.Errorf("parse failed")
}

func (client *productClient) describeBackendByOneId(backendID string) (*ulb.ULBBackendSet, string, string, error) {
	lbID, listenerID, err := client.getLBIdAndVServerIdFromBackendId(backendID)
	if err != nil {
		return nil, "", "", err
	}
	vserverSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return nil, "", "", err
	}
	for i := range vserverSet.BackendSet {
		backend := vserverSet.BackendSet[i]
		if backend.BackendId == backendID {
			return &backend, lbID, listenerID, nil
		}
	}
	return nil, "", "", newNotFoundError(getNotFoundMessage("backend", backendID))
}

func (client *productClient) getLBIdAndVServerIdFromBackendId(backendID string) (string, string, error) {
	request := client.ulbconn.NewDescribeULBRequest()
	var ulbSets []ulb.ULBSet
	const limit = 100
	for offset := 0; ; offset += limit {
		request.Limit = ucloud.Int(limit)
		request.Offset = ucloud.Int(offset)
		response, err := client.ulbconn.DescribeULB(request)
		if err != nil {
			return "", "", err
		}
		if response == nil || len(response.DataSet) < 1 {
			break
		}
		ulbSets = append(ulbSets, response.DataSet...)
		for _, lb := range ulbSets {
			for _, listener := range lb.VServerSet {
				for _, backend := range listener.BackendSet {
					if backend.BackendId == backendID {
						return lb.ULBId, listener.VServerId, nil
					}
				}
			}
		}
		if len(response.DataSet) < limit {
			break
		}
	}
	return "", "", fmt.Errorf("parse failed")
}

func (client *productClient) describePolicyByOneId(policyID string) (*ulb.ULBPolicySet, string, string, error) {
	lbID, listenerID, err := client.getLBIdAndVServerIdFromPolicyId(policyID)
	if err != nil {
		return nil, "", "", err
	}
	vserverSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return nil, "", "", err
	}
	for i := range vserverSet.PolicySet {
		policy := vserverSet.PolicySet[i]
		if policy.PolicyId == policyID {
			return &policy, lbID, listenerID, nil
		}
	}
	return nil, "", "", newNotFoundError(getNotFoundMessage("policy", policyID))
}

func (client *productClient) getLBIdAndVServerIdFromPolicyId(policyID string) (string, string, error) {
	request := client.ulbconn.NewDescribeULBRequest()
	var ulbSets []ulb.ULBSet
	const limit = 100
	for offset := 0; ; offset += limit {
		request.Limit = ucloud.Int(limit)
		request.Offset = ucloud.Int(offset)
		response, err := client.ulbconn.DescribeULB(request)
		if err != nil {
			return "", "", err
		}
		if response == nil || len(response.DataSet) < 1 {
			break
		}
		ulbSets = append(ulbSets, response.DataSet...)
		for _, lb := range ulbSets {
			for _, listener := range lb.VServerSet {
				for _, policy := range listener.PolicySet {
					if policy.PolicyId == policyID {
						return lb.ULBId, listener.VServerId, nil
					}
				}
			}
		}
		if len(response.DataSet) < limit {
			break
		}
	}
	return "", "", fmt.Errorf("parse failed")
}

func (client *productClient) describeFirewallByIdAndType(resourceID, resourceType string) (*unet.FirewallDataSet, error) {
	request := client.unetconn.NewDescribeFirewallRequest()
	request.ResourceId = ucloud.String(resourceID)
	request.ResourceType = ucloud.String(resourceType)
	response, err := client.unetconn.DescribeFirewall(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 54002 {
			return nil, newNotFoundError("")
		}
		return nil, err
	}
	if response == nil || len(response.DataSet) < 1 {
		return nil, newNotFoundError("")
	}
	return &response.DataSet[0], nil
}
