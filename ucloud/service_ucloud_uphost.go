package ucloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"github.com/ucloud/ucloud-sdk-go/services/uphost"
	"github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

func (client *UCloudClient) describeBareMetalInstanceById(instanceId string) (*uphost.PHostSet, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	req := client.uphostconn.NewDescribePHostRequest()
	req.PHostId = []string{instanceId}

	resp, err := client.uphostconn.DescribePHost(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 16001 {
			return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
		}
		return nil, err
	}
	if len(resp.PHostSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("instance", instanceId))
	}

	return &resp.PHostSet[0], nil
}

func (client *UCloudClient) getRaidTypeById(instanceId string) (string, error) {
	if instanceId == "" {
		return "", newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	req := client.genericClient.NewGenericRequest()
	err := req.SetPayload(map[string]interface{}{
		"Action":  "DescribePHost",
		"PHostId": []string{instanceId},
	})
	if err != nil {
		return "", errors.New("failed to set payload")
	}
	resp, err := client.genericClient.GenericInvoke(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 16001 {
			return "", newNotFoundError(getNotFoundMessage("instance", instanceId))
		}
		return "", err
	}
	respJson, err := json.Marshal(resp.GetPayload())
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}
	phostSetResult := gjson.Get(string(respJson), "PHostSet")
	if !phostSetResult.Exists() || len(phostSetResult.Array()) == 0 {
		return "", newNotFoundError(getNotFoundMessage("instance", instanceId))
	}
	phostResult := phostSetResult.Array()[0]
	if phostResult.Get("PhostClass").String() == "LocalDisk" {
		diskSetResult := phostResult.Get("DiskSet")
		if diskSetResult.Exists() {
			for _, disk := range diskSetResult.Array() {
				raid := disk.Get("Raid").String()
				if raid != "" {
					return raid, nil
				}
			}
		}
	}

	return "", nil
}

type DescribePHostImageResponse struct {
	response.CommonBase

	// 镜像列表 PHostImageSet
	ImageSet []PHostImageSet

	// 满足条件的镜像总数
	TotalCount int
}

/*
PHostImageSet - DescribePHostImage
*/
type PHostImageSet struct {
	// 镜像描述
	ImageDescription string

	// 镜像ID
	ImageId string

	// 镜像名称
	ImageName string

	// 裸金属2.0参数。镜像大小。
	ImageSize int

	// 枚举值：Base=>基础镜像，Custom=>自制镜像。
	ImageType string

	// 操作系统名称
	OsName string

	// 操作系统类型
	OsType string

	// 裸金属2.0参数。镜像当前状态。
	State string

	// 支持的机型
	Support []string

	// 当前版本
	Version string
}

func (client *UCloudClient) getBareMetalImages(imageReq uphost.DescribePHostImageRequest) (*DescribePHostImageResponse, error) {
	req := client.genericClient.NewGenericRequest()
	payload, _ := json.Marshal(imageReq)
	payLoadMap := make(map[string]interface{})
	err := json.Unmarshal(payload, &payLoadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
	}
	payLoadMap["Action"] = "DescribePHostImage"
	for k, v := range payLoadMap {
		if v == nil {
			delete(payLoadMap, k)
		}
	}
	err = req.SetPayload(payLoadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to set payload: %v", err)
	}
	resp, err := client.genericClient.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	imageResp := &DescribePHostImageResponse{}
	err = mapstructure.Decode(resp.GetPayload(), imageResp)
	if err != nil {
		return nil, fmt.Errorf("fail to decode response: %v", err)
	}

	return imageResp, nil
}
