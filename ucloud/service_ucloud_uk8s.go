package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func (client *UCloudClient) describeUK8sClusterById(instanceId string) (*uk8s.ClusterSet, error) {
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

func (client *UCloudClient) describeUK8sClusterNodeById(instanceId string) ([]uk8s.NodeInfoV2, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceId))
	}
	req := client.uk8sconn.NewListUK8SClusterNodeV2Request()
	req.ClusterId = ucloud.String(instanceId)

	resp, err := client.uk8sconn.ListUK8SClusterNodeV2(req)
	if err != nil {
		return nil, err
	}
	if len(resp.NodeSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("uk8s_cluster", instanceId))
	}

	return resp.NodeSet, nil
}
