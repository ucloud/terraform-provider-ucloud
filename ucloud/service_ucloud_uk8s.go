package ucloud

import (
	"fmt"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

// resolveUK8SMachine resolves the cpu/mem(MB)/machineType for a uk8s master or
// node from the new dedicated fields (cpu/memory/machine_type). It falls back to
// parsing the deprecated instance_type string when machine_type is not set, so
// existing configurations keep working. The get callback maps a schema key
// (already prefixed by the caller, e.g. "master.0.cpu") to its value, allowing
// the same logic to serve both *schema.ResourceData and *schema.ResourceDiff.
func resolveUK8SMachine(get func(string) interface{}, cpuKey, memKey, machineTypeKey, instanceTypeKey string) (cpu, memMB int, machineType string, err error) {
	if mt, ok := get(machineTypeKey).(string); ok && mt != "" {
		machineType = strings.ToUpper(mt)
		cpu, _ = get(cpuKey).(int)
		memMB, _ = get(memKey).(int)
		if cpu <= 0 {
			return 0, 0, "", fmt.Errorf("%q must be set and greater than 0 when %q is set", cpuKey, machineTypeKey)
		}
		if memMB <= 0 {
			return 0, 0, "", fmt.Errorf("%q must be set and greater than 0 when %q is set", memKey, machineTypeKey)
		}
		return cpu, memMB, machineType, nil
	}

	if it, ok := get(instanceTypeKey).(string); ok && it != "" {
		t, err := parseInstanceType(it)
		if err != nil {
			return 0, 0, "", err
		}
		return t.CPU, t.Memory, strings.ToUpper(t.HostType), nil
	}

	return 0, 0, "", fmt.Errorf("one of %q or %q must be set", machineTypeKey, instanceTypeKey)
}

// uk8sErrorWithMessage compensates for a bug in some uk8s SDK response structs
// (e.g. AddUK8SUHostNodeResponse, DelUK8SClusterNodeV2Response) that redeclare a
// `Message` field, shadowing response.CommonBase.Message. Because the SDK builds
// its error from CommonBase.Message via GetMessage(), the real reason returned by
// the backend (e.g. "resource not enough") lands in the shadowing field and is
// lost from err. This appends that field back when it carries extra information.
func uk8sErrorWithMessage(err error, respMessage string) error {
	if err == nil {
		return nil
	}
	if respMessage != "" && !strings.Contains(err.Error(), respMessage) {
		return fmt.Errorf("%s %s", err, respMessage)
	}
	return err
}

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
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("error on reading uk8s_cluster %q, %s", instanceId, resp.GetMessage())
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
