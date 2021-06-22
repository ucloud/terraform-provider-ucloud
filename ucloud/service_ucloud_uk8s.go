package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *UCloudClient) describeUK8SClusterById(instanceId string) (*uk8s.ClusterSet, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceId))
	}
	req := client.uk8sconn.NewListUK8SClusterV2Request()
	req.ClusterId = ucloud.String(instanceId)

	resp, err := client.uk8sconn.ListUK8SClusterV2(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ClusterSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceId))
	}

	return &resp.ClusterSet[0], nil
}

func (client *UCloudClient) describeUK8SClusterNodeById(instanceId string) ([]uk8s.NodeInfoV2, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceId))
	}
	req := client.uk8sconn.NewListUK8SClusterNodeV2Request()
	req.ClusterId = ucloud.String(instanceId)

	resp, err := client.uk8sconn.ListUK8SClusterNodeV2(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 94007 {
			return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceId))
		}
		return nil, err
	}
	if len(resp.NodeSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_node", instanceId))
	}

	return resp.NodeSet, nil
}

func (client *UCloudClient) describeUK8SClusterNodeByResourceId(instanceId, resourceId string) (*uk8s.NodeInfoV2, error) {
	if resourceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_node", resourceId))
	}

	nodes, err := client.describeUK8SClusterNodeById(instanceId)
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if node.NodeId == resourceId {
			return &node, nil
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("uk8s_node", resourceId))
}
