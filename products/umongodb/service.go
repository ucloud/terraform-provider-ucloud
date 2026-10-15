package umongodb

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

const (
	deleteReplSetAction = "DeleteUMongoDBReplSet"
	deleteShardedAction = "DeleteUMongoDBShardedCluster"
)

type createReplSetResponse struct {
	response.CommonBase
	ClusterId string
}

type createShardedResponse struct {
	response.CommonBase
	ShardedClusterId string
}

type deleteClusterRequest struct {
	request.CommonBase
	ClusterId        *string
	ShardedClusterId *string
}

type actionResponse struct {
	response.CommonBase
	Message string
}

func deleteIdentity(kind string) (action, field string) {
	if kind == kindSharded {
		return deleteShardedAction, "ShardedClusterId"
	}
	return deleteReplSetAction, "ClusterId"
}

func describeCluster(client *umongodb.UMongoDBClient, id, clusterType, zone string) (*umongodb.ClusterInfo, error) {
	if id == "" {
		return nil, newNotFoundError(fmt.Sprintf("the specified umongodb cluster %s is not found", id))
	}

	req := client.NewDescribeUMongoDBInstanceRequest()
	req.ClusterId = ucloud.String(id)
	if clusterType != "" {
		req.ClusterType = ucloud.String(clusterType)
	}
	if zone != "" {
		req.Zone = ucloud.String(zone)
	}

	resp, err := client.DescribeUMongoDBInstance(req)
	if err != nil {
		return nil, fmt.Errorf("DescribeUMongoDBInstance %s: %w", id, err)
	}
	if resp == nil || resp.ClusterInfo.ClusterId == "" || resp.ClusterInfo.State == apiStateDeleted {
		return nil, newNotFoundError(fmt.Sprintf("the specified umongodb cluster %s is not found", id))
	}
	info := resp.ClusterInfo
	return &info, nil
}

func createReplSet(client *umongodb.UMongoDBClient, data *schema.ResourceData) (string, error) {
	req := client.NewCreateUMongoDBReplSetRequest()
	req.Zone = ucloud.String(data.Get("availability_zone").(string))
	req.Name = ucloud.String(data.Get("name").(string))
	req.DBVersion = ucloud.String(data.Get("db_version").(string))
	req.AdminPassword = ucloud.String(data.Get("admin_password").(string))
	req.DiskSpace = ucloud.Int(data.Get("disk_space").(int))
	req.MachineTypeId = ucloud.String(data.Get("machine_type_id").(string))
	req.NodeCount = ucloud.Int(data.Get("node_count").(int))
	req.ChargeType = ucloud.String(chargeTypeAPI(data.Get("charge_type").(string)))
	req.Quantity = ucloud.Int(data.Get("quantity").(int))
	if value, ok := data.GetOk("listen_port"); ok {
		req.ListenPort = ucloud.Int(value.(int))
	}
	if value, ok := data.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	}
	if zones := expandStringList(data.Get("cross_zones")); len(zones) > 0 {
		req.CrossZones = zones
	}
	if labels := expandReplSetLabels(data.Get("label")); len(labels) > 0 {
		req.Labels = labels
	}
	if groups := expandReplSetSecGroups(data.Get("sec_group")); len(groups) > 0 {
		req.SecGroupId = groups
	}

	var resp createReplSetResponse
	if err := client.InvokeAction("CreateUMongoDBReplSet", req, &resp); err != nil {
		return "", fmt.Errorf("CreateUMongoDBReplSet: %w", err)
	}
	if resp.ClusterId == "" {
		return "", fmt.Errorf("CreateUMongoDBReplSet returned no ClusterId")
	}
	return resp.ClusterId, nil
}

func createShardedCluster(client *umongodb.UMongoDBClient, data *schema.ResourceData) (string, error) {
	req := client.NewCreateUMongoDBShardedClusterRequest()
	req.Zone = ucloud.String(data.Get("availability_zone").(string))
	req.Name = ucloud.String(data.Get("name").(string))
	req.DBVersion = ucloud.String(data.Get("db_version").(string))
	req.AdminPassword = ucloud.String(data.Get("admin_password").(string))
	req.DiskSpace = ucloud.Int(data.Get("disk_space").(int))
	req.MachineTypeId = ucloud.String(data.Get("machine_type_id").(string))
	req.NodeCount = ucloud.Int(data.Get("node_count").(int))
	req.MongosNodeCount = ucloud.Int(data.Get("mongos_node_count").(int))
	req.ShardCount = ucloud.Int(data.Get("shard_count").(int))
	req.ChargeType = ucloud.String(chargeTypeAPI(data.Get("charge_type").(string)))
	req.Quantity = ucloud.Int(data.Get("quantity").(int))
	if value, ok := data.GetOk("listen_port"); ok {
		req.ListenPort = ucloud.Int(value.(int))
	}
	if value, ok := data.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("mongos_machine_type_id"); ok {
		req.MongosMachineTypeId = ucloud.String(value.(string))
	}
	if labels := expandShardedLabels(data.Get("label")); len(labels) > 0 {
		req.Labels = labels
	}
	if groups := expandShardedSecGroups(data.Get("sec_group")); len(groups) > 0 {
		req.SecGroupId = groups
	}

	var resp createShardedResponse
	if err := client.InvokeAction("CreateUMongoDBShardedCluster", req, &resp); err != nil {
		return "", fmt.Errorf("CreateUMongoDBShardedCluster: %w", err)
	}
	if resp.ShardedClusterId == "" {
		return "", fmt.Errorf("CreateUMongoDBShardedCluster returned no ShardedClusterId")
	}
	return resp.ShardedClusterId, nil
}

func deleteCluster(client *umongodb.UMongoDBClient, kind, id, zone string) error {
	action, _ := deleteIdentity(kind)
	req := &deleteClusterRequest{}
	client.SetupRequest(req)
	req.Zone = ucloud.String(zone)
	if kind == kindSharded {
		req.ShardedClusterId = ucloud.String(id)
	} else {
		req.ClusterId = ucloud.String(id)
	}
	var resp actionResponse
	if err := client.InvokeAction(action, req, &resp); err != nil {
		return fmt.Errorf("%s %s: %w", action, id, err)
	}
	return nil
}

func modifyClusterName(client *umongodb.UMongoDBClient, id, zone, name string) error {
	req := client.NewModifyUMongoDBAttributeRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	req.Name = ucloud.String(name)
	if _, err := client.ModifyUMongoDBAttribute(req); err != nil {
		return fmt.Errorf("ModifyUMongoDBAttribute %s: %w", id, err)
	}
	return nil
}

func modifyAdminPassword(client *umongodb.UMongoDBClient, id, zone, password string) error {
	req := client.NewModifyUMongoDBAdminPasswordRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	req.Password = ucloud.String(base64.RawURLEncoding.EncodeToString([]byte(password)))
	if _, err := client.ModifyUMongoDBAdminPassword(req); err != nil {
		return fmt.Errorf("ModifyUMongoDBAdminPassword %s: %w", id, err)
	}
	return nil
}

func resizeCluster(client *umongodb.UMongoDBClient, id string, diskSpace int, machineTypeID, mongosMachineTypeID string) error {
	req := client.NewResizeUMongoDBInstanceRequest()
	req.ClusterId = ucloud.String(id)
	if diskSpace > 0 {
		req.DiskSpace = ucloud.Int(diskSpace)
	}
	if machineTypeID != "" {
		req.MachineTypeId = ucloud.String(machineTypeID)
	}
	if mongosMachineTypeID != "" {
		req.MongosMachineTypeId = ucloud.String(mongosMachineTypeID)
	}
	if _, err := client.ResizeUMongoDBInstance(req); err != nil {
		return fmt.Errorf("ResizeUMongoDBInstance %s: %w", id, err)
	}
	return nil
}

func modifyBackupParam(client *umongodb.UMongoDBClient, id, zone string, data *schema.ResourceData) error {
	week := data.Get("auto_backup_week").(string)
	backupTime := data.Get("auto_backup_time").(string)
	disabled := data.Get("backup_disabled").(bool)
	if !disabled && (week == "" || backupTime == "") {
		param, err := readBackupParam(client, id, zone)
		if err != nil {
			return err
		}
		if week == "" {
			week = param.AutoBackupToggleWeek
		}
		if backupTime == "" {
			backupTime = param.AutoBackupToggleTime
		}
	}
	week, backupTime, sendDisabled, err := backupParamUpdate(week, backupTime, disabled, data.HasChange("backup_disabled"))
	if err != nil {
		return err
	}
	req := client.NewModifyUMongoDBBackupParamRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	if week != "" {
		req.AutoBackupToggleWeek = ucloud.String(week)
	}
	if backupTime != "" {
		req.AutoBackupToggleTime = ucloud.String(backupTime)
	}
	if sendDisabled != nil {
		req.Disabled = ucloud.Bool(*sendDisabled)
	}
	if req.AutoBackupToggleWeek == nil && req.AutoBackupToggleTime == nil && req.Disabled == nil {
		return nil
	}
	if _, err := client.ModifyUMongoDBBackupParam(req); err != nil {
		return fmt.Errorf("ModifyUMongoDBBackupParam %s: %w", id, err)
	}
	return nil
}

// backupParamUpdate matches ModifyBackupParam: disabling does not store the window,
// and any other change must send both weekday and time.
func backupParamUpdate(week, backupTime string, disabled, disabledChanged bool) (string, string, *bool, error) {
	if disabled {
		value := true
		return "", "", &value, nil
	}
	if week == "" || backupTime == "" {
		return "", "", nil, fmt.Errorf("auto_backup_week and auto_backup_time must both be set unless backup_disabled is true")
	}
	if disabledChanged {
		value := false
		return week, backupTime, &value, nil
	}
	return week, backupTime, nil, nil
}

func readBackupParam(client *umongodb.UMongoDBClient, id, zone string) (*umongodb.BackupParam, error) {
	req := client.NewGetUMongoDBBackupParamRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	resp, err := client.GetUMongoDBBackupParam(req)
	if err != nil {
		return nil, fmt.Errorf("GetUMongoDBBackupParam %s: %w", id, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("GetUMongoDBBackupParam %s returned no backup policy", id)
	}
	param := resp.DataSet
	return &param, nil
}

func startCluster(client *umongodb.UMongoDBClient, id, zone string) error {
	req := client.NewStartUMongoDBClusterRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	if _, err := client.StartUMongoDBCluster(req); err != nil {
		return fmt.Errorf("StartUMongoDBCluster %s: %w", id, err)
	}
	return nil
}

func stopCluster(client *umongodb.UMongoDBClient, id, zone string) error {
	req := client.NewStopUMongoDBClusterRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	if _, err := client.StopUMongoDBCluster(req); err != nil {
		return fmt.Errorf("StopUMongoDBCluster %s: %w", id, err)
	}
	return nil
}

func restartCluster(client *umongodb.UMongoDBClient, id, zone string) error {
	req := client.NewRestartUMongoDBClusterRequest()
	req.ClusterId = ucloud.String(id)
	req.Zone = ucloud.String(zone)
	if _, err := client.RestartUMongoDBCluster(req); err != nil {
		return fmt.Errorf("RestartUMongoDBCluster %s: %w", id, err)
	}
	return nil
}

func createBackup(client *umongodb.UMongoDBClient, clusterID, zone, name string) (string, error) {
	req := client.NewBackupUMongoDBClusterRequest()
	req.SetRetryable(false)
	req.ClusterId = ucloud.String(clusterID)
	req.Zone = ucloud.String(zone)
	req.BackupName = ucloud.String(name)
	resp, err := client.BackupUMongoDBCluster(req)
	if err != nil {
		return "", fmt.Errorf("BackupUMongoDBCluster %s: %w", clusterID, err)
	}
	if resp == nil || resp.BackupId == "" {
		return "", fmt.Errorf("BackupUMongoDBCluster %s returned no BackupId", clusterID)
	}
	return resp.BackupId, nil
}

func findBackup(client *umongodb.UMongoDBClient, clusterID, zone, backupID, backupName string, startTime int) (*umongodb.BackupInfo, error) {
	backups, err := listBackups(client, clusterID, zone)
	if err != nil {
		return nil, err
	}
	if info := matchBackup(backups, backupID, "", 0); info != nil {
		return info, nil
	}
	if info := matchListedBackup(backups, backupName, startTime); info != nil {
		return info, nil
	}
	return nil, newNotFoundError(fmt.Sprintf("the specified umongodb backup %s is not found", backupID))
}

func findCreatedBackup(client *umongodb.UMongoDBClient, clusterID, zone, backupID, backupName string, createdAfter int) (*umongodb.BackupInfo, error) {
	backups, err := listBackups(client, clusterID, zone)
	if err != nil {
		return nil, err
	}
	if info := matchBackup(backups, backupID, backupName, createdAfter); info != nil {
		return info, nil
	}
	return nil, newNotFoundError(fmt.Sprintf("the specified umongodb backup %s is not found", backupID))
}

func listBackups(client *umongodb.UMongoDBClient, clusterID, zone string) ([]umongodb.BackupInfo, error) {
	req := client.NewListUMongoDBBackupRequest()
	req.ClusterId = ucloud.String(clusterID)
	req.Zone = ucloud.String(zone)
	resp, err := client.ListUMongoDBBackup(req)
	if err != nil {
		return nil, fmt.Errorf("ListUMongoDBBackup %s: %w", clusterID, err)
	}
	if resp == nil {
		return nil, nil
	}
	return resp.DataSet, nil
}

// matchBackup finds the create response id. A sharded snapshot list collapses a batch
// onto one BackupId, which can differ from the id returned by BackupUMongoDBCluster.
// The same backup name is used only when its start time is not older than createdAfter.
func matchBackup(backups []umongodb.BackupInfo, backupID, backupName string, createdAfter int) *umongodb.BackupInfo {
	for index := range backups {
		if backupID != "" && backups[index].BackupId == backupID {
			info := backups[index]
			return &info
		}
	}
	if backupName == "" {
		return nil
	}
	var matched *umongodb.BackupInfo
	matches := 0
	for index := range backups {
		item := backups[index]
		if item.BackupName != backupName || item.StartTime < createdAfter {
			continue
		}
		matches++
		copy := item
		matched = &copy
	}
	if matches != 1 {
		return nil
	}
	return matched
}

// matchListedBackup follows a sharded snapshot after the list folds the batch
// onto a different BackupId. The folded record keeps the name and the earliest start time.
func matchListedBackup(backups []umongodb.BackupInfo, backupName string, startTime int) *umongodb.BackupInfo {
	if backupName == "" || startTime == 0 {
		return nil
	}
	var matched *umongodb.BackupInfo
	matches := 0
	for index := range backups {
		item := backups[index]
		if item.BackupName != backupName || item.StartTime != startTime {
			continue
		}
		matches++
		copy := item
		matched = &copy
	}
	if matches != 1 {
		return nil
	}
	return matched
}

func clusterDeleteSettled(info *umongodb.ClusterInfo) bool {
	if info == nil {
		return false
	}
	switch info.State {
	case "InitFailed", "StartFailed":
	default:
		return false
	}
	seen := false
	if info.ConfigReplicaInfo.ReplicaId != "" {
		seen = true
		if info.ConfigReplicaInfo.State != apiStateDeleted {
			return false
		}
	}
	for _, replica := range info.DataReplicaInfos {
		seen = true
		if replica.State != apiStateDeleted {
			return false
		}
	}
	for _, node := range info.MongosInfo {
		seen = true
		if node.State != apiStateDeleted {
			return false
		}
	}
	return seen
}

func listInstances(client *umongodb.UMongoDBClient, clusterID, zone string) ([]umongodb.MongodbInstance, error) {
	req := client.NewListUMongoDBInstancesRequest()
	if clusterID != "" {
		req.ClusterId = ucloud.String(clusterID)
	}
	if zone != "" {
		req.Zone = ucloud.String(zone)
	}
	resp, err := client.ListUMongoDBInstances(req)
	if err != nil {
		return nil, fmt.Errorf("ListUMongoDBInstances: %w", err)
	}
	if resp == nil {
		return nil, nil
	}
	return resp.DataSet, nil
}

func waitUntilGone(client *umongodb.UMongoDBClient, id, clusterType, zone string, timeout time.Duration) error {
	conf := &resource.StateChangeConf{
		Pending:    []string{"Deleting"},
		Target:     []string{"gone"},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,
		Refresh: func() (interface{}, string, error) {
			info, err := describeCluster(client, id, clusterType, zone)
			if err != nil {
				if isNotFoundError(err) {
					return info, "gone", nil
				}
				return nil, "", err
			}
			if clusterDeleteSettled(info) {
				return info, "gone", nil
			}
			if isFailedState(info.State) {
				return info, "", fmt.Errorf("umongodb cluster %s entered %s", id, info.State)
			}
			return info, "Deleting", nil
		},
	}
	if _, err := conf.WaitForState(); err != nil {
		return fmt.Errorf("wait for umongodb cluster %s to be deleted: %w", id, err)
	}
	return nil
}

func waitCluster(client *umongodb.UMongoDBClient, id, clusterType, zone string, pending, target []string, timeout time.Duration) error {
	pendingSet := stringSet(pending)
	targetSet := stringSet(target)
	conf := &resource.StateChangeConf{
		Pending:    append([]string{}, pending...),
		Target:     append([]string{}, target...),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 15 * time.Second,
		Refresh: func() (interface{}, string, error) {
			info, err := describeCluster(client, id, clusterType, zone)
			if err != nil {
				if isNotFoundError(err) {
					if _, ok := targetSet["gone"]; ok {
						return info, "gone", nil
					}
					return nil, pending[0], nil
				}
				return nil, "", err
			}
			if _, ok := targetSet[info.State]; ok {
				return info, info.State, nil
			}
			if _, ok := pendingSet[info.State]; ok || info.State == "" {
				if info.State == "" {
					return info, pending[0], nil
				}
				return info, info.State, nil
			}
			if isFailedState(info.State) {
				return info, "", fmt.Errorf("umongodb cluster %s entered %s", id, info.State)
			}
			return info, info.State, nil
		},
	}
	_, err := conf.WaitForState()
	if err != nil {
		return fmt.Errorf("wait for umongodb cluster %s: %w", id, err)
	}
	return nil
}

func isFailedState(state string) bool {
	switch state {
	case "InitFailed", "StartFailed", "StopFailed", "RestartFailed", "UpgradeFailed", "DeleteFailed":
		return true
	default:
		return false
	}
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func runningPending() []string {
	return []string{"Initing", "Starting", "Restarting", "Upgrading", "Switching"}
}

func encodeAdminPassword(password string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(password))
}
