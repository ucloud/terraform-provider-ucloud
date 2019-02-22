package ucloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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

func (client *UCloudClient) describeDBParameterGroupByIdAndZone(paramGroupId, zone string) (*udb.UDBParamGroupSet, error) {
	req := client.udbconn.NewDescribeUDBParamGroupRequest()
	req.Zone = ucloud.String(zone)
	pgId, err := strconv.Atoi(paramGroupId)
	if err != nil {
		return nil, fmt.Errorf("transform param group id %q to int failed, %s", paramGroupId, err)
	}
	req.GroupId = ucloud.Int(pgId)

	resp, err := client.udbconn.DescribeUDBParamGroup(req)
	if err != nil {
		if uErr, ok := err.(uerr.Error); ok && uErr.Code() == 7011 {
			return nil, newNotFoundError(getNotFoundMessage("db_param_group", paramGroupId))
		}
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("db_param_group", paramGroupId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) describeDBBackupByIdAndZone(backupId, zone string) (*udb.UDBBackupSet, error) {
	req := client.udbconn.NewDescribeUDBBackupRequest()
	req.Zone = ucloud.String(zone)
	buId, err := strconv.Atoi(backupId)
	if err != nil {
		return nil, fmt.Errorf("transform backup id %q to int failed, %s", backupId, err)
	}
	req.BackupId = ucloud.Int(buId)

	resp, err := client.udbconn.DescribeUDBBackup(req)
	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("db_backup", backupId))
	}

	return &resp.DataSet[0], nil
}

func (client *UCloudClient) dbWaitForState(dbId string, target []string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     target,
		Timeout:    5 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {
			db, err := client.describeDBInstanceById(dbId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			state := db.State
			if state == "RecoverFail" {
				return nil, "", fmt.Errorf("db instance recover failed, please make sure your %q is correct and matched with the other parameters", "backup_id")
			}

			if !isStringIn(state, target) {
				state = statusPending
			}

			return db, state, nil
		},
	}
}
