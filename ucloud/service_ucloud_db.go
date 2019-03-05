package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func (client *UCloudClient) describeDBInstanceById(dbInstanceId string) (*udb.UDBInstanceSet, error) {
	req := client.udbconn.NewDescribeUDBInstanceRequest()
	req.DBId = ucloud.String(dbInstanceId)

	resp, err := client.udbconn.DescribeUDBInstance(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 230 {
			return nil, newNotFoundError(getNotFoundMessage("db_instance", dbInstanceId))
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("db_instance", dbInstanceId))
	}

	return &resp.DataSet[0], nil
}
