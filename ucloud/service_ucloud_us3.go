package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *UCloudClient) describeUS3BucketById(instanceId string) (*ufile.UFileBucketSet, error) {
	if instanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("us3_bucket", instanceId))
	}
	req := client.us3conn.NewDescribeBucketRequest()
	req.BucketName = ucloud.String(instanceId)

	resp, err := client.us3conn.DescribeBucket(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 15010 {
			return nil, newNotFoundError(getNotFoundMessage("us3_bucket", instanceId))
		}
		return nil, err
	}
	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("us3_bucket", instanceId))
	}

	return &resp.DataSet[0], nil
}
