package umongodb

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
)

func resourceUCloudUMongoDBReplSet() *schema.Resource {
	return &schema.Resource{
		Create:        resourceUCloudUMongoDBReplSetCreate,
		Read:          resourceUCloudUMongoDBReplSetRead,
		Update:        resourceUCloudUMongoDBReplSetUpdate,
		Delete:        resourceUCloudUMongoDBReplSetDelete,
		Importer:      clusterImporter(),
		Timeouts:      clusterTimeouts(),
		CustomizeDiff: clusterCustomizeDiff(),
		Schema:        baseClusterSchema(),
	}
}

func resourceUCloudUMongoDBShardedCluster() *schema.Resource {
	return &schema.Resource{
		Create:        resourceUCloudUMongoDBShardedClusterCreate,
		Read:          resourceUCloudUMongoDBShardedClusterRead,
		Update:        resourceUCloudUMongoDBShardedClusterUpdate,
		Delete:        resourceUCloudUMongoDBShardedClusterDelete,
		Importer:      clusterImporter(),
		Timeouts:      clusterTimeouts(),
		CustomizeDiff: clusterCustomizeDiff(),
		Schema:        shardedClusterSchema(),
	}
}

// Stopping or starting replaces the computed state and connection string.
// The legacy SDK drops those new values unless the plan marks them unknown.
func clusterCustomizeDiff() schema.CustomizeDiffFunc {
	return customdiff.All(
		customdiff.ComputedIf("state", statusChanging),
		customdiff.ComputedIf("connect_url", statusChanging),
	)
}

func statusChanging(diff *schema.ResourceDiff, _ interface{}) bool {
	return diff.HasChange("status")
}

func resourceUCloudUMongoDBReplSetCreate(data *schema.ResourceData, meta interface{}) error {
	return createAndWait(data, meta, kindReplica)
}

func resourceUCloudUMongoDBShardedClusterCreate(data *schema.ResourceData, meta interface{}) error {
	return createAndWait(data, meta, kindSharded)
}

func resourceUCloudUMongoDBReplSetRead(data *schema.ResourceData, meta interface{}) error {
	return readClusterResource(data, meta, kindReplica)
}

func resourceUCloudUMongoDBShardedClusterRead(data *schema.ResourceData, meta interface{}) error {
	return readClusterResource(data, meta, kindSharded)
}

func resourceUCloudUMongoDBReplSetUpdate(data *schema.ResourceData, meta interface{}) error {
	return updateClusterResource(data, meta, kindReplica)
}

func resourceUCloudUMongoDBShardedClusterUpdate(data *schema.ResourceData, meta interface{}) error {
	return updateClusterResource(data, meta, kindSharded)
}

func resourceUCloudUMongoDBReplSetDelete(data *schema.ResourceData, meta interface{}) error {
	return deleteClusterResource(data, meta, kindReplica)
}

func resourceUCloudUMongoDBShardedClusterDelete(data *schema.ResourceData, meta interface{}) error {
	return deleteClusterResource(data, meta, kindSharded)
}

func createAndWait(data *schema.ResourceData, meta interface{}, kind string) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating umongodb %s, %s", kind, err)
	}

	var id string
	if kind == kindSharded {
		id, err = createShardedCluster(client, data)
	} else {
		id, err = createReplSet(client, data)
	}
	if err != nil {
		return err
	}
	data.SetId(id)

	zone := data.Get("availability_zone").(string)
	if err := waitCluster(client, id, clusterTypeFor(kind), zone, runningPending(), []string{apiStateRunning}, data.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}
	if err := applyConfiguredBackup(client, data, id, zone); err != nil {
		return err
	}
	if data.Get("status").(string) == statusStopped {
		if err := stopCluster(client, id, zone); err != nil {
			return err
		}
		if err := waitCluster(client, id, clusterTypeFor(kind), zone, stopPending(), []string{apiStateStopped}, data.Timeout(schema.TimeoutCreate)); err != nil {
			return err
		}
	}
	return readClusterResource(data, meta, kind)
}

func updateClusterResource(data *schema.ResourceData, meta interface{}, kind string) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating umongodb %s %q, %s", kind, data.Id(), err)
	}
	id := data.Id()
	zone := data.Get("availability_zone").(string)
	clusterType := clusterTypeFor(kind)
	timeout := data.Timeout(schema.TimeoutUpdate)

	if data.HasChange("name") && !data.IsNewResource() {
		if err := modifyClusterName(client, id, zone, data.Get("name").(string)); err != nil {
			return err
		}
	}
	resizeMongos := kind == kindSharded && data.HasChange("mongos_machine_type_id")
	if !data.IsNewResource() && (data.HasChange("disk_space") || data.HasChange("machine_type_id") || resizeMongos) {
		mongosType := ""
		if resizeMongos {
			mongosType = data.Get("mongos_machine_type_id").(string)
		}
		if err := resizeCluster(client, id, data.Get("disk_space").(int), data.Get("machine_type_id").(string), mongosType); err != nil {
			return err
		}
		if err := waitCluster(client, id, clusterType, zone, []string{"Upgrading"}, []string{apiStateRunning, apiStateStopped}, timeout); err != nil {
			return err
		}
	}
	if !data.IsNewResource() && (data.HasChange("auto_backup_week") || data.HasChange("auto_backup_time") || data.HasChange("backup_disabled")) {
		if err := modifyBackupParam(client, id, zone, data); err != nil {
			return err
		}
	}
	passwordChange := !data.IsNewResource() && data.HasChange("admin_password")
	desiredRunning := data.Get("status").(string) == statusRunning
	oldTrigger, newTrigger := data.GetChange("restart_trigger")
	restartChange := shouldRestart(oldTrigger.(string), newTrigger.(string), data.Get("status").(string), data.IsNewResource())
	startedForUpdate := false
	if (passwordChange || restartChange) && data.Get("state").(string) != apiStateRunning {
		if err := startCluster(client, id, zone); err != nil {
			return err
		}
		if err := waitCluster(client, id, clusterType, zone, startPending(), []string{apiStateRunning}, timeout); err != nil {
			return err
		}
		startedForUpdate = true
	}
	if passwordChange {
		if err := modifyAdminPassword(client, id, zone, data.Get("admin_password").(string)); err != nil {
			return err
		}
	}
	if data.HasChange("status") && !data.IsNewResource() {
		switch data.Get("status").(string) {
		case statusStopped:
			if err := stopCluster(client, id, zone); err != nil {
				return err
			}
			if err := waitCluster(client, id, clusterType, zone, stopPending(), []string{apiStateStopped}, timeout); err != nil {
				return err
			}
		case statusRunning:
			if err := startCluster(client, id, zone); err != nil {
				return err
			}
			if err := waitCluster(client, id, clusterType, zone, startPending(), []string{apiStateRunning}, timeout); err != nil {
				return err
			}
		}
	} else if startedForUpdate && !desiredRunning {
		if err := stopCluster(client, id, zone); err != nil {
			return err
		}
		if err := waitCluster(client, id, clusterType, zone, stopPending(), []string{apiStateStopped}, timeout); err != nil {
			return err
		}
	}
	if restartChange {
		if err := restartCluster(client, id, zone); err != nil {
			return err
		}
		if err := waitCluster(client, id, clusterType, zone, restartPending(), []string{apiStateRunning}, timeout); err != nil {
			return err
		}
	}
	return readClusterResource(data, meta, kind)
}

func shouldRestart(oldTrigger, newTrigger, status string, isNew bool) bool {
	return !isNew && oldTrigger != "" && oldTrigger != newTrigger && status == statusRunning
}

func deleteClusterResource(data *schema.ResourceData, meta interface{}, kind string) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting umongodb %s %q, %s", kind, data.Id(), err)
	}
	id := data.Id()
	zone := data.Get("availability_zone").(string)
	info, err := describeCluster(client, id, clusterTypeFor(kind), zone)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error on reading umongodb %s %q before delete, %s", kind, id, err)
	}
	timeout := data.Timeout(schema.TimeoutDelete)
	switch deleteStepForState(info.State) {
	case deleteStepGone:
		return nil
	case deleteStepWait:
		return waitUntilGone(client, id, clusterTypeFor(kind), zone, timeout)
	case deleteStepWaitStop:
		if err := waitCluster(client, id, clusterTypeFor(kind), zone, stopPending(), []string{apiStateStopped}, timeout); err != nil {
			return err
		}
	case deleteStepStop:
		if err := stopCluster(client, id, zone); err != nil {
			return err
		}
		if err := waitCluster(client, id, clusterTypeFor(kind), zone, stopPending(), []string{apiStateStopped}, timeout); err != nil {
			return err
		}
	case deleteStepNow:
	default:
		return fmt.Errorf("umongodb %s %q is %s and cannot be deleted", kind, id, info.State)
	}
	if err := deleteCluster(client, kind, id, zone); err != nil {
		if _, readErr := describeCluster(client, id, clusterTypeFor(kind), zone); isNotFoundError(readErr) {
			return nil
		}
		return err
	}
	return waitUntilGone(client, id, clusterTypeFor(kind), zone, data.Timeout(schema.TimeoutDelete))
}

func readClusterResource(data *schema.ResourceData, meta interface{}, kind string) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading umongodb %s %q, %s", kind, data.Id(), err)
	}
	zone := data.Get("availability_zone").(string)
	info, err := describeCluster(client, data.Id(), clusterTypeFor(kind), zone)
	if err != nil {
		if isNotFoundError(err) {
			data.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading umongodb %s %q, %s", kind, data.Id(), err)
	}
	setClusterFields(data, info, kind)
	if backupConfigured(data) {
		param, paramErr := readBackupParam(client, data.Id(), zone)
		if paramErr != nil {
			return paramErr
		}
		setBackupFields(data, param)
	}
	return nil
}

func backupConfigured(data *schema.ResourceData) bool {
	return data.Get("auto_backup_week").(string) != "" || data.Get("auto_backup_time").(string) != "" || data.Get("backup_disabled").(bool)
}

func applyConfiguredBackup(client *umongodb.UMongoDBClient, data *schema.ResourceData, id, zone string) error {
	if !backupConfigured(data) {
		return nil
	}
	return modifyBackupParam(client, id, zone, data)
}

func setClusterFields(data *schema.ResourceData, info *umongodb.ClusterInfo, kind string) {
	_ = data.Set("name", info.InstanceName)
	_ = data.Set("db_version", info.DBVersion)
	_ = data.Set("disk_space", info.DiskSpace)
	_ = data.Set("machine_type_id", info.MachineTypeId)
	_ = data.Set("availability_zone", info.Zone)
	setReturnedString(data, "subnet_id", info.SubnetId)
	setReturnedString(data, "vpc_id", info.VPCId)
	setReturnedString(data, "tag", info.Tag)
	_ = data.Set("state", info.State)
	_ = data.Set("connect_url", info.ConnectURL)
	_ = data.Set("create_time", info.CreateTime)
	switch info.State {
	case apiStateRunning:
		_ = data.Set("status", statusRunning)
	case apiStateStopped:
		_ = data.Set("status", statusStopped)
	}
	if kind != kindSharded && len(info.CrossZones) > 0 {
		_ = data.Set("cross_zones", info.CrossZones)
	}
	if kind == kindSharded {
		_ = data.Set("shard_count", info.ShardCount)
		_ = data.Set("mongos_node_count", info.MongosCount)
		if info.ShardNodeCount > 0 {
			_ = data.Set("node_count", info.ShardNodeCount)
		}
		return
	}
	if count := replSetNodeCount(info); count > 0 {
		_ = data.Set("node_count", count)
	}
}

func setBackupFields(data *schema.ResourceData, param *umongodb.BackupParam) {
	if param == nil {
		return
	}
	if param.Disabled {
		if data.Get("backup_disabled").(bool) {
			_ = data.Set("backup_disabled", true)
		}
		return
	}
	if data.Get("backup_disabled").(bool) {
		_ = data.Set("backup_disabled", false)
	}
	if data.Get("auto_backup_week").(string) != "" {
		_ = data.Set("auto_backup_week", param.AutoBackupToggleWeek)
	}
	if data.Get("auto_backup_time").(string) != "" {
		_ = data.Set("auto_backup_time", param.AutoBackupToggleTime)
	}
}

func setReturnedString(data *schema.ResourceData, key, value string) {
	if value == "" {
		return
	}
	_ = data.Set(key, value)
}

const (
	deleteStepReject = iota
	deleteStepGone
	deleteStepWait
	deleteStepWaitStop
	deleteStepStop
	deleteStepNow
)

func deleteStepForState(state string) int {
	switch state {
	case apiStateDeleted:
		return deleteStepGone
	case "Deleting":
		return deleteStepWait
	case "Stopping":
		return deleteStepWaitStop
	case apiStateRunning, "StopFailed":
		return deleteStepStop
	case apiStateStopped, "InitFailed", "StartFailed", "DeleteFailed":
		return deleteStepNow
	default:
		return deleteStepReject
	}
}

func stopPending() []string {
	return []string{apiStateRunning, "Stopping"}
}

func startPending() []string {
	return []string{apiStateStopped, "Starting"}
}

func restartPending() []string {
	return []string{"Restarting", "Stopping", "Starting"}
}

func clusterTypeFor(kind string) string {
	if kind == kindSharded {
		return clusterTypeSharded
	}
	return clusterTypeReplica
}
