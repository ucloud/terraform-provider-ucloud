package sqlserver

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

// describeDBInstanceByID returns one instance by DBId, or a not-found error.
func describeDBInstanceByID(client *ucloud.Client, dbId, zone string) (map[string]interface{}, error) {
	if dbId == "" {
		return nil, newNotFoundError(getNotFoundMessage("sqlserver_instance", dbId))
	}

	payload := basePayload(client, zone)
	payload["ClassType"] = "sqlserver"
	payload["DBId"] = dbId

	resp, err := invoke(client, "DescribeUDBInstance", payload)
	if err != nil {
		return nil, err
	}

	dataSet, ok := resp["DataSet"].([]interface{})
	if !ok || len(dataSet) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("sqlserver_instance", dbId))
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
		payload["ClassType"] = "sqlserver"
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

// sqlserverInstanceStateRefreshFunc drives a StateChangeConf: it reports the
// current instance state, mapping any non-target state to pending.
func sqlserverInstanceStateRefreshFunc(client *ucloud.Client, dbId, zone string, target []string) resource.StateRefreshFunc {
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
			if state == sqlserverStateFail || state == sqlserverStateRecoverFail ||
				state == sqlserverStateStartFail || state == sqlserverStateShutdownFail ||
				state == sqlserverStateUpgradeFail || state == sqlserverStateDeleteFail {
				return nil, "", fmt.Errorf("sqlserver instance %q is in state %q", dbId, state)
			}
			state = statusPending
		}
		return db, state, nil
	}
}
