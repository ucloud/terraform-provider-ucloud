package uk8s

import (
	"fmt"

	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func describeUK8SClusterById(client *sdkuk8s.UK8SClient, instanceID string) (*sdkuk8s.ClusterSet, error) {
	if instanceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceID))
	}

	req := client.NewListUK8SClusterV2Request()
	req.ClusterId = ucloud.String(instanceID)

	resp, err := client.ListUK8SClusterV2(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading uk8s_cluster %q, %s", instanceID, resp.GetMessage())
	}
	if len(resp.ClusterSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceID))
	}

	return &resp.ClusterSet[0], nil
}

func describeUK8SClusterNodeById(client *sdkuk8s.UK8SClient, instanceID string) ([]sdkuk8s.NodeInfoV2, error) {
	if instanceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceID))
	}

	req := client.NewListUK8SClusterNodeV2Request()
	req.ClusterId = ucloud.String(instanceID)

	resp, err := client.ListUK8SClusterNodeV2(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 94007 {
			return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceID))
		}
		return nil, err
	}
	if len(resp.NodeSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_node", instanceID))
	}

	return resp.NodeSet, nil
}

func describeUK8SClusterNodeByResourceId(client *sdkuk8s.UK8SClient, instanceID, resourceID string) (*sdkuk8s.NodeInfoV2, error) {
	if resourceID == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_node", resourceID))
	}

	nodes, err := describeUK8SClusterNodeById(client, instanceID)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.NodeId == resourceID {
			return &node, nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("uk8s_node", resourceID))
}
