package ucloud

import (
	"encoding/base64"
	"github.com/ucloud/ucloud-sdk-go/services/cube"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"gopkg.in/yaml.v2"
)

type provider struct {
	ChargeType     string `yaml:"chargeType"`
	ContainerCount string `yaml:"containerCount"`
	CpuPlatform    string `yaml:"cpuPlatform"`
	Region         string `yaml:"region"`
	SubnetId       string `yaml:"subnetId"`
	VpcId          string `yaml:"vpcId"`
	Zone           string `yaml:"zone"`
}

type metadata struct {
	CreateTimestamp string   `yaml:"creationTimestamp"`
	Group           string   `yaml:"group"`
	Provider        provider `yaml:"provider"`
}

type status struct {
	Phase string `yaml:"phase"`
	PodIp string `yaml:"podIp"`
}

type cubePodInfo struct {
	Metadata metadata `yaml:"metadata"`
	Status   status   `yaml:"status"`
}
type cubePodExtendInfo struct {
	CubePodInfo    cubePodInfo
	CubeExtendInfo cube.CubeExtendInfo
}

func (client *UCloudClient) describeCubePodById(instanceId string) (*cubePodExtendInfo, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("cube_pod", instanceId))
	}
	req := client.cubeconn.NewGetCubeExtendInfoRequest()
	req.CubeIds = ucloud.String(instanceId)

	resp, err := client.cubeconn.GetCubeExtendInfo(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ExtendInfo) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("cube_pod", instanceId))
	}

	reqPod := client.cubeconn.NewGetCubePodRequest()
	reqPod.CubeId = ucloud.String(instanceId)

	respPod, err := client.cubeconn.GetCubePod(reqPod)
	if err != nil {
		return nil, err
	}

	if resp.ExtendInfo[0].Expiration == 0 && respPod.Pod == "" {
		return nil, newNotFoundError(getNotFoundMessage("cube_pod", instanceId))
	}

	podByte, err := base64.StdEncoding.DecodeString(respPod.Pod)
	if err != nil {
		return nil, err
	}

	var podInfo cubePodInfo
	if err := yaml.Unmarshal(podByte, &podInfo); err != nil {
		return nil, err
	}

	return &cubePodExtendInfo{
		podInfo,
		resp.ExtendInfo[0],
	}, nil
}
