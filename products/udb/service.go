package udb

import (
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func describeDBInstanceByID(client *udb.UDBClient, dbInstanceId string) (*udb.UDBInstanceSet, error) {
	if dbInstanceId == "" {
		return nil, newNotFoundError(getNotFoundMessage("db_instance", dbInstanceId))
	}
	req := client.NewDescribeUDBInstanceRequest()
	req.DBId = ucloud.String(dbInstanceId)

	resp, err := client.DescribeUDBInstance(req)
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
