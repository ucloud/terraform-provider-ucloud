package udb_mysql

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

// describeDBInstanceByID returns one instance by DBId, or a not-found error.
func describeDBInstanceByID(client *ucloud.Client, dbId, zone string) (map[string]interface{}, error) {
	if dbId == "" {
		return nil, newNotFoundError(getNotFoundMessage("udb_mysql_instance", dbId))
	}

	payload := basePayload(client, zone)
	payload["ClassType"] = "sql"
	payload["DBId"] = dbId

	resp, err := invoke(client, "DescribeUDBInstance", payload)
	if err != nil {
		return nil, err
	}

	dataSet, ok := resp["DataSet"].([]interface{})
	if !ok || len(dataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("udb_mysql_instance", dbId))
	}
	item, ok := dataSet[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected describe response item type %T", dataSet[0])
	}
	return item, nil
}

// describeDBInstances pages through DescribeUDBInstance with the given filters
// (e.g. VPCId) and returns every matched instance.
func describeDBInstances(client *ucloud.Client, zone string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	const limit = 100
	offset := 0

	for {
		payload := basePayload(client, zone)
		payload["ClassType"] = "sql"
		payload["Limit"] = limit
		payload["Offset"] = offset
		for k, v := range filters {
			payload[k] = v
		}

		resp, err := invoke(client, "DescribeUDBInstance", payload)
		if err != nil {
			return nil, err
		}

		dataSet, _ := resp["DataSet"].([]interface{})
		for _, item := range dataSet {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		if len(dataSet) < limit {
			break
		}
		offset += limit
	}
	return result, nil
}

// getDefaultParamGroupID resolves the default parameter template for a DB
// version via ListUDBParamTemplate (TemplateType omitted == default).
func getDefaultParamGroupID(client *ucloud.Client, zone, dbVersion string) (int, error) {
	payload := basePayload(client, zone)
	payload["DBVersion"] = dbVersion

	resp, err := invoke(client, "ListUDBParamTemplate", payload)
	if err != nil {
		return 0, err
	}
	dataSet, _ := resp["DataSet"].([]interface{})
	if len(dataSet) == 0 {
		return 0, fmt.Errorf("no default param template found for version %s", dbVersion)
	}
	item, ok := dataSet[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected param template item type %T", dataSet[0])
	}
	return payloadInt(item, "Id"), nil
}

// mysqlInstanceStateRefreshFunc drives a StateChangeConf: it reports the
// current instance state, mapping any non-target state to pending.
func mysqlInstanceStateRefreshFunc(client *ucloud.Client, dbId, zone string, target []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		db, err := describeDBInstanceByID(client, dbId, zone)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		state := payloadString(db, "State")
		if !isStringIn(state, target) {
			if state == mysqlStateFail || state == mysqlStateRecoverFail ||
				state == mysqlStateStartFail || state == mysqlStateShutdownFail ||
				state == mysqlStateUpgradeFail || state == mysqlStateDeleteFail {
				return nil, "", fmt.Errorf("mysql instance %q is in state %q", dbId, state)
			}
			state = statusPending
		}
		return db, state, nil
	}
}
