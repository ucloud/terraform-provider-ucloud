// Code is generated by ucloud-model, DO NOT EDIT IT.

package ufs

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

// UFS API Schema

// AddUFSVolumeMountPointRequest is request schema for AddUFSVolumeMountPoint action
type AddUFSVolumeMountPointRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](https://docs.ucloud.cn/api/summary/get_project_list)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](https://docs.ucloud.cn/api/summary/regionlist)
	// Region *string `required:"true"`

	// 挂载点名称
	MountPointName *string `required:"true"`

	// Subnet ID
	SubnetId *string `required:"true"`

	// 文件系统ID
	VolumeId *string `required:"true"`

	// Vpc ID
	VpcId *string `required:"true"`
}

// AddUFSVolumeMountPointResponse is response schema for AddUFSVolumeMountPoint action
type AddUFSVolumeMountPointResponse struct {
	response.CommonBase
}

// NewAddUFSVolumeMountPointRequest will create request of AddUFSVolumeMountPoint action.
func (c *UFSClient) NewAddUFSVolumeMountPointRequest() *AddUFSVolumeMountPointRequest {
	req := &AddUFSVolumeMountPointRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)
	return req
}

/*
API: AddUFSVolumeMountPoint

添加文件系统挂载点
*/
func (c *UFSClient) AddUFSVolumeMountPoint(req *AddUFSVolumeMountPointRequest) (*AddUFSVolumeMountPointResponse, error) {
	var err error
	var res AddUFSVolumeMountPointResponse

	reqCopier := *req

	err = c.Client.InvokeAction("AddUFSVolumeMountPoint", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// CreateUFSVolumeRequest is request schema for CreateUFSVolume action
type CreateUFSVolumeRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// 计费模式，枚举值为： Year，按年付费； Month，按月付费； Dynamic，按需付费（需开启权限）； Trial，试用（需开启权限） 默认为Dynamic
	ChargeType *string `required:"false"`

	// 使用的代金券id
	CouponId *string `required:"false"`

	// 文件系统协议，枚举值，NFSv3表示NFS V3协议，NFSv4表示NFS V4协议
	ProtocolType *string `required:"true"`

	// 购买时长 默认: 1
	Quantity *int `required:"false"`

	// 备注
	Remark *string `required:"false"`

	// 文件系统大小，单位为GB，最大不超过20T，香港容量型必须为100的整数倍，Size最小为500GB，北京，上海，广州的容量型必须为1024的整数倍，Size最小为1024GB。性能型文件系统Size最小为100GB
	Size *int `required:"true"`

	// 文件系统存储类型，枚举值，Basic表示容量型，Advanced表示性能型
	StorageType *string `required:"true"`

	// 文件系统所属业务组
	Tag *string `required:"false"`

	// 文件系统名称
	VolumeName *string `required:"false"`
}

// CreateUFSVolumeResponse is response schema for CreateUFSVolume action
type CreateUFSVolumeResponse struct {
	response.CommonBase

	// 文件系统ID
	VolumeId string

	// 文件系统名称
	VolumeName string

	// 文件系统挂载点状态
	VolumeStatus string
}

// NewCreateUFSVolumeRequest will create request of CreateUFSVolume action.
func (c *UFSClient) NewCreateUFSVolumeRequest() *CreateUFSVolumeRequest {
	req := &CreateUFSVolumeRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(false)
	return req
}

/*
API: CreateUFSVolume

创建文件系统
*/
func (c *UFSClient) CreateUFSVolume(req *CreateUFSVolumeRequest) (*CreateUFSVolumeResponse, error) {
	var err error
	var res CreateUFSVolumeResponse

	reqCopier := *req

	err = c.Client.InvokeAction("CreateUFSVolume", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// DescribeUFSVolume2Request is request schema for DescribeUFSVolume2 action
type DescribeUFSVolume2Request struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// 文件列表长度
	Limit *int `required:"false"`

	// 文件列表起始
	Offset *int `required:"false"`

	// 文件系统ID
	VolumeId *string `required:"false"`
}

// DescribeUFSVolume2Response is response schema for DescribeUFSVolume2 action
type DescribeUFSVolume2Response struct {
	response.CommonBase

	// 文件系统详细信息列表
	DataSet []UFSVolumeInfo2

	// 文件系统总数
	TotalCount int
}

// NewDescribeUFSVolume2Request will create request of DescribeUFSVolume2 action.
func (c *UFSClient) NewDescribeUFSVolume2Request() *DescribeUFSVolume2Request {
	req := &DescribeUFSVolume2Request{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

/*
API: DescribeUFSVolume2

获取文件系统列表
*/
func (c *UFSClient) DescribeUFSVolume2(req *DescribeUFSVolume2Request) (*DescribeUFSVolume2Response, error) {
	var err error
	var res DescribeUFSVolume2Response

	reqCopier := *req

	err = c.Client.InvokeAction("DescribeUFSVolume2", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// DescribeUFSVolumeMountpointRequest is request schema for DescribeUFSVolumeMountpoint action
type DescribeUFSVolumeMountpointRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](https://docs.ucloud.cn/api/summary/get_project_list)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](https://docs.ucloud.cn/api/summary/regionlist)
	// Region *string `required:"true"`

	// 文件系统ID
	VolumeId *string `required:"true"`
}

// DescribeUFSVolumeMountpointResponse is response schema for DescribeUFSVolumeMountpoint action
type DescribeUFSVolumeMountpointResponse struct {
	response.CommonBase

	//
	DataSet []MountPointInfo

	// 文件系统能创建的最大挂载点数目
	MaxMountPointNum int

	// 目前的挂载点总数
	TotalMountPointNum int
}

// NewDescribeUFSVolumeMountpointRequest will create request of DescribeUFSVolumeMountpoint action.
func (c *UFSClient) NewDescribeUFSVolumeMountpointRequest() *DescribeUFSVolumeMountpointRequest {
	req := &DescribeUFSVolumeMountpointRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

/*
API: DescribeUFSVolumeMountpoint

获取文件系统挂载点信息
*/
func (c *UFSClient) DescribeUFSVolumeMountpoint(req *DescribeUFSVolumeMountpointRequest) (*DescribeUFSVolumeMountpointResponse, error) {
	var err error
	var res DescribeUFSVolumeMountpointResponse

	reqCopier := *req

	err = c.Client.InvokeAction("DescribeUFSVolumeMountpoint", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// ExtendUFSVolumeRequest is request schema for ExtendUFSVolume action
type ExtendUFSVolumeRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// 文件系统大小，单位为GB，最大不超过20T，香港容量型必须为100的整数倍，Size最小为500GB，北京，上海，广州的容量型必须为1024的整数倍，Size最小为1024GB。性能型文件系统Size最小为100GB
	Size *int `required:"true"`

	// 文件系统ID
	VolumeId *string `required:"true"`
}

// ExtendUFSVolumeResponse is response schema for ExtendUFSVolume action
type ExtendUFSVolumeResponse struct {
	response.CommonBase
}

// NewExtendUFSVolumeRequest will create request of ExtendUFSVolume action.
func (c *UFSClient) NewExtendUFSVolumeRequest() *ExtendUFSVolumeRequest {
	req := &ExtendUFSVolumeRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

/*
API: ExtendUFSVolume

文件系统扩容
*/
func (c *UFSClient) ExtendUFSVolume(req *ExtendUFSVolumeRequest) (*ExtendUFSVolumeResponse, error) {
	var err error
	var res ExtendUFSVolumeResponse

	reqCopier := *req

	err = c.Client.InvokeAction("ExtendUFSVolume", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// RemoveUFSVolumeRequest is request schema for RemoveUFSVolume action
type RemoveUFSVolumeRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	// Region *string `required:"true"`

	// 文件系统ID
	VolumeId *string `required:"true"`
}

// RemoveUFSVolumeResponse is response schema for RemoveUFSVolume action
type RemoveUFSVolumeResponse struct {
	response.CommonBase
}

// NewRemoveUFSVolumeRequest will create request of RemoveUFSVolume action.
func (c *UFSClient) NewRemoveUFSVolumeRequest() *RemoveUFSVolumeRequest {
	req := &RemoveUFSVolumeRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

/*
API: RemoveUFSVolume

删除UFS文件系统
*/
func (c *UFSClient) RemoveUFSVolume(req *RemoveUFSVolumeRequest) (*RemoveUFSVolumeResponse, error) {
	var err error
	var res RemoveUFSVolumeResponse

	reqCopier := *req

	err = c.Client.InvokeAction("RemoveUFSVolume", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// RemoveUFSVolumeMountPointRequest is request schema for RemoveUFSVolumeMountPoint action
type RemoveUFSVolumeMountPointRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](https://docs.ucloud.cn/api/summary/get_project_list)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](https://docs.ucloud.cn/api/summary/regionlist)
	// Region *string `required:"true"`

	// Subnet ID
	SubnetId *string `required:"true"`

	// 文件系统ID
	VolumeId *string `required:"true"`

	// Vpc ID
	VpcId *string `required:"true"`
}

// RemoveUFSVolumeMountPointResponse is response schema for RemoveUFSVolumeMountPoint action
type RemoveUFSVolumeMountPointResponse struct {
	response.CommonBase
}

// NewRemoveUFSVolumeMountPointRequest will create request of RemoveUFSVolumeMountPoint action.
func (c *UFSClient) NewRemoveUFSVolumeMountPointRequest() *RemoveUFSVolumeMountPointRequest {
	req := &RemoveUFSVolumeMountPointRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

/*
API: RemoveUFSVolumeMountPoint

删除文件系统挂载点
*/
func (c *UFSClient) RemoveUFSVolumeMountPoint(req *RemoveUFSVolumeMountPointRequest) (*RemoveUFSVolumeMountPointResponse, error) {
	var err error
	var res RemoveUFSVolumeMountPointResponse

	reqCopier := *req

	err = c.Client.InvokeAction("RemoveUFSVolumeMountPoint", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}

// UpdateUFSVolumeInfoRequest is request schema for UpdateUFSVolumeInfo action
type UpdateUFSVolumeInfoRequest struct {
	request.CommonBase

	// [公共参数] 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](https://docs.ucloud.cn/api/summary/get_project_list)
	// ProjectId *string `required:"false"`

	// [公共参数] 地域。 参见 [地域和可用区列表](https://docs.ucloud.cn/api/summary/regionlist)
	// Region *string `required:"true"`

	// 文件系统备注（文件系统名称／备注至少传入其中一个）
	Remark *string `required:"false"`

	// 文件系统ID
	VolumeId *string `required:"true"`

	// 文件系统名称（文件系统名称／备注至少传入其中一个）
	VolumeName *string `required:"false"`
}

// UpdateUFSVolumeInfoResponse is response schema for UpdateUFSVolumeInfo action
type UpdateUFSVolumeInfoResponse struct {
	response.CommonBase
}

// NewUpdateUFSVolumeInfoRequest will create request of UpdateUFSVolumeInfo action.
func (c *UFSClient) NewUpdateUFSVolumeInfoRequest() *UpdateUFSVolumeInfoRequest {
	req := &UpdateUFSVolumeInfoRequest{}

	// setup request with client config
	c.Client.SetupRequest(req)

	// setup retryable with default retry policy (retry for non-create action and common error)
	req.SetRetryable(true)
	return req
}

/*
API: UpdateUFSVolumeInfo

更改文件系统相关信息（名称／备注）
*/
func (c *UFSClient) UpdateUFSVolumeInfo(req *UpdateUFSVolumeInfoRequest) (*UpdateUFSVolumeInfoResponse, error) {
	var err error
	var res UpdateUFSVolumeInfoResponse

	reqCopier := *req

	err = c.Client.InvokeAction("UpdateUFSVolumeInfo", &reqCopier, &res)
	if err != nil {
		return &res, err
	}

	return &res, nil
}
