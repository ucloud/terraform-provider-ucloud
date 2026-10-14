package umongodb

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New())); err != nil {
		t.Fatalf("register UMongoDB product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UMongoDB product: %v", err)
	}
	for _, name := range []string{
		"ucloud_umongodb_replicaset",
		"ucloud_umongodb_sharded_cluster",
		"ucloud_umongodb_backup",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Fatalf("%s is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_umongodb_instance", "ucloud_umongodb_instances"} {
		if provider.DataSourcesMap[name] == nil {
			t.Fatalf("%s is not registered", name)
		}
	}
}

func TestClusterSchemaContract(t *testing.T) {
	replset := New().Registration().Resources["ucloud_umongodb_replicaset"]
	sharded := New().Registration().Resources["ucloud_umongodb_sharded_cluster"]
	backup := New().Registration().Resources["ucloud_umongodb_backup"]
	for _, resource := range []*schema.Resource{replset, sharded} {
		if resource.Create == nil || resource.Read == nil || resource.Update == nil || resource.Delete == nil {
			t.Fatal("cluster CRUD callbacks are incomplete")
		}
		if resource.Importer == nil || resource.Importer.State == nil {
			t.Fatal("cluster importer is missing")
		}
		if resource.CustomizeDiff == nil {
			t.Fatal("cluster status changes must mark state and connect_url computed")
		}
		if !resource.Schema["admin_password"].Sensitive {
			t.Fatal("admin_password must be sensitive")
		}
		if !resource.Schema["node_count"].ForceNew {
			t.Fatal("node_count must be ForceNew")
		}
		if resource.Schema["disk_space"].ForceNew {
			t.Fatal("disk_space is resized in place")
		}
	}
	if sharded.Schema["shard_count"] == nil || !sharded.Schema["shard_count"].ForceNew {
		t.Fatal("sharded cluster shard_count must be ForceNew")
	}
	if backup.Delete == nil || backup.Update != nil {
		t.Fatal("backup resource should create and delete state only")
	}
}

func TestCreateResponseKeepsClusterID(t *testing.T) {
	var replset createReplSetResponse
	if err := json.Unmarshal([]byte(`{"RetCode":0,"ClusterId":"umongodb-rs-1"}`), &replset); err != nil {
		t.Fatal(err)
	}
	if replset.ClusterId != "umongodb-rs-1" {
		t.Fatalf("replica set id = %q", replset.ClusterId)
	}

	var sharded createShardedResponse
	if err := json.Unmarshal([]byte(`{"RetCode":0,"ShardedClusterId":"umongodb-sc-1"}`), &sharded); err != nil {
		t.Fatal(err)
	}
	if sharded.ShardedClusterId != "umongodb-sc-1" {
		t.Fatalf("sharded id = %q", sharded.ShardedClusterId)
	}
}

func TestDeleteIdentity(t *testing.T) {
	action, field := deleteIdentity(kindReplica)
	if action != deleteReplSetAction || field != "ClusterId" {
		t.Fatalf("replica delete = %s %s", action, field)
	}
	action, field = deleteIdentity(kindSharded)
	if action != deleteShardedAction || field != "ShardedClusterId" {
		t.Fatalf("sharded delete = %s %s", action, field)
	}
}

func TestEncodeAdminPasswordUsesRawURL(t *testing.T) {
	encoded := encodeAdminPassword("secret?")
	if strings.Contains(encoded, "=") || strings.Contains(encoded, "+") || strings.Contains(encoded, "/") {
		t.Fatalf("password encoding = %q, want raw url base64", encoded)
	}
	if encoded == "secret?" {
		t.Fatal("password was not encoded")
	}
}

func TestDescribeClusterEmptyIDIsNotFound(t *testing.T) {
	client := newClient(&ucloud.Config{Region: "cn-bj2"}, &auth.Credential{}, nil).(*umongodb.UMongoDBClient)
	if _, err := describeCluster(client, "", clusterTypeReplica, "cn-bj2-05"); err == nil || !isNotFoundError(err) {
		t.Fatalf("describe empty id error = %v, want not found", err)
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("clientFromMeta accepted a non-runtime value")
	}
}

func TestClientUsesUMongoDBRuntime(t *testing.T) {
	config := &ucloud.Config{Region: "cn-bj2", Timeout: time.Second}
	client, ok := newClient(config, &auth.Credential{}, nil).(*umongodb.UMongoDBClient)
	if !ok {
		t.Fatalf("newClient returned %T", client)
	}
	if got := client.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("timeout = %s", got)
	}
	if got := providerRegion(client); got != "cn-bj2" {
		t.Fatalf("region = %q", got)
	}
}

func TestDiskSpaceValidation(t *testing.T) {
	for value, wantErr := range map[int]bool{10: true, 20: false, 25: true, 32000: false, 32010: true} {
		_, errors := validateDiskSpace(value, "disk_space")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("disk_space %d err=%v, wantErr=%t", value, errors, wantErr)
		}
	}
}

func TestChargeTypeAPI(t *testing.T) {
	if got := chargeTypeAPI("month"); got != "Month" {
		t.Fatalf("charge type = %q", got)
	}
}

func TestShardedSchemaOmitsUnsupportedFields(t *testing.T) {
	sharded := New().Registration().Resources["ucloud_umongodb_sharded_cluster"].Schema
	if _, ok := sharded["cross_zones"]; ok {
		t.Fatal("sharded cluster exposes cross_zones, which CreateUMongoDBShardedCluster does not accept")
	}
	if _, ok := sharded["template_id"]; ok {
		t.Fatal("sharded cluster exposes template_id; config templates are out of scope")
	}
	if _, ok := New().Registration().Resources["ucloud_umongodb_backup"].Schema["cluster_type"]; ok {
		t.Fatal("backup cluster_type is not sent to BackupUMongoDBCluster")
	}
}

func TestDeleteStepForState(t *testing.T) {
	cases := map[string]int{
		"Running":       deleteStepStop,
		"StopFailed":    deleteStepStop,
		"Stopping":      deleteStepWaitStop,
		"Stopped":       deleteStepNow,
		"InitFailed":    deleteStepNow,
		"StartFailed":   deleteStepNow,
		"DeleteFailed":  deleteStepNow,
		"Deleting":      deleteStepWait,
		"Deleted":       deleteStepGone,
		"Restarting":    deleteStepReject,
		"UpgradeFailed": deleteStepReject,
	}
	for state, want := range cases {
		if got := deleteStepForState(state); got != want {
			t.Errorf("state %s step = %d, want %d", state, got, want)
		}
	}
	for _, state := range restartPending() {
		if state != "Starting" && state != "Stopping" && state != "Restarting" {
			t.Fatalf("unexpected restart pending state %s", state)
		}
	}
}

func TestBackupParamUpdate(t *testing.T) {
	week, window, disabled, err := backupParamUpdate("1,2", "1:00~2:00", true, true)
	if err != nil || week != "" || window != "" || disabled == nil || !*disabled {
		t.Fatalf("disable update = %q %q %v %v", week, window, disabled, err)
	}
	if _, _, _, err := backupParamUpdate("1,2", "", false, false); err == nil {
		t.Fatal("a single backup window field was accepted")
	}
	week, window, disabled, err = backupParamUpdate("1,2", "1:00~2:00", false, false)
	if err != nil || week != "1,2" || window != "1:00~2:00" || disabled != nil {
		t.Fatalf("window update = %q %q %v %v", week, window, disabled, err)
	}
}

func TestDisabledBackupKeepsConfiguredWindow(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceUCloudUMongoDBReplSet().Schema, map[string]interface{}{
		"availability_zone": "cn-bj2-05",
		"name":              "tf-name1",
		"db_version":        "MongoDB 4.2",
		"admin_password":    "secret1",
		"disk_space":        20,
		"machine_type_id":   "o.mongo2m.medium",
		"node_count":        3,
		"auto_backup_week":  "1,2",
		"auto_backup_time":  "1:00~2:00",
		"backup_disabled":   true,
	})
	setBackupFields(data, &umongodb.BackupParam{
		Disabled:             true,
		AutoBackupToggleWeek: "3,7",
		AutoBackupToggleTime: "5:00~6:00",
	})
	if got := data.Get("auto_backup_week").(string); got != "1,2" {
		t.Fatalf("week overwritten with %q", got)
	}
	if got := data.Get("backup_disabled").(bool); !got {
		t.Fatal("backup_disabled was cleared")
	}
}

func TestReadSetsReturnedIdentityFields(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceUCloudUMongoDBReplSet().Schema, map[string]interface{}{
		"availability_zone": "cn-bj2-05",
		"name":              "tf-name1",
		"db_version":        "MongoDB 4.2",
		"admin_password":    "secret1",
		"disk_space":        20,
		"machine_type_id":   "o.mongo2m.medium",
		"node_count":        3,
	})
	setClusterFields(data, &umongodb.ClusterInfo{
		InstanceName:  "tf-name1",
		DBVersion:     "MongoDB 4.2",
		DiskSpace:     20,
		MachineTypeId: "o.mongo2m.medium",
		Zone:          "cn-bj2-05",
		VPCId:         "uvnet-1",
		SubnetId:      "subnet-1",
		Tag:           "biz",
		State:         apiStateRunning,
		CrossZones:    []string{"cn-bj2-02"},
	}, kindReplica)
	if data.Get("vpc_id").(string) != "uvnet-1" || data.Get("subnet_id").(string) != "subnet-1" || data.Get("tag").(string) != "biz" {
		t.Fatalf("identity fields = vpc %q subnet %q tag %q", data.Get("vpc_id"), data.Get("subnet_id"), data.Get("tag"))
	}
	if len(expandStringList(data.Get("cross_zones"))) != 1 {
		t.Fatalf("cross_zones = %#v", data.Get("cross_zones"))
	}
}

func TestBackupImportID(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceUCloudUMongoDBBackup().Schema, map[string]interface{}{
		"availability_zone": "cn-bj2-05",
		"cluster_id":        "placeholder",
		"backup_name":       "backup",
	})
	data.SetId("cn-bj2-05/umongodb-rs-1/bak-1")
	imported, err := resourceUCloudUMongoDBBackupImport(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if imported[0].Id() != "bak-1" || imported[0].Get("cluster_id").(string) != "umongodb-rs-1" {
		t.Fatalf("imported id=%s cluster=%s", imported[0].Id(), imported[0].Get("cluster_id"))
	}
	if !backupCreateFinished("Backup_Stopped") || !backupCreateFinished("Backup_Success") {
		t.Fatal("terminal backup states were not recognized")
	}
}

func TestListFilterAndStableID(t *testing.T) {
	instances := filterInstances([]umongodb.MongodbInstance{
		{ClusterId: "b", Zone: "cn-bj2-05"},
		{ClusterId: "a", Zone: "cn-bj2-02"},
		{ClusterId: "c", Zone: "cn-bj2-05"},
	}, "", "cn-bj2-05")
	if len(instances) != 2 {
		t.Fatalf("filtered = %d", len(instances))
	}
	if got := dataSourceID([]string{"b", "a"}); got != "a,b" {
		t.Fatalf("data source id = %q", got)
	}
}

func TestReviewFixes(t *testing.T) {
	replset := New().Registration().Resources["ucloud_umongodb_replicaset"].Schema
	if replset["cross_zones"].Type != schema.TypeSet {
		t.Fatal("cross_zones must be a set so describe order cannot replace the cluster")
	}
	if !replset["charge_type"].DiffSuppressFunc("charge_type", "", "month", nil) {
		t.Fatal("empty imported charge_type should match the month default")
	}
	if replset["charge_type"].DiffSuppressFunc("charge_type", "", "year", nil) {
		t.Fatal("empty imported charge_type must not hide a year charge")
	}
	sharded := New().Registration().Resources["ucloud_umongodb_sharded_cluster"].Schema
	if _, errors := sharded["shard_count"].ValidateFunc(1, "shard_count"); len(errors) == 0 {
		t.Fatal("shard_count 1 was accepted")
	}
	if _, errors := sharded["shard_count"].ValidateFunc(10, "shard_count"); len(errors) == 0 {
		t.Fatal("shard_count 10 was accepted")
	}
	if _, errors := sharded["shard_count"].ValidateFunc(9, "shard_count"); len(errors) != 0 {
		t.Fatal("shard_count 9 was rejected")
	}
	if _, errors := sharded["mongos_node_count"].ValidateFunc(6, "mongos_node_count"); len(errors) == 0 {
		t.Fatal("mongos_node_count 6 was accepted")
	}
	if _, errors := sharded["mongos_node_count"].ValidateFunc(5, "mongos_node_count"); len(errors) != 0 {
		t.Fatal("mongos_node_count 5 was rejected")
	}
	for _, state := range stopPending() {
		if state == "StopFailed" {
			t.Fatal("StopFailed is a terminal stop failure, not a pending state")
		}
	}
	instance := New().Registration().DataSources["ucloud_umongodb_instance"].Schema
	if !instance["availability_zone"].Computed || !instance["cluster_type"].Computed {
		t.Fatal("instance data source must compute zone and cluster type when they are omitted")
	}

	backups := []umongodb.BackupInfo{
		{BackupId: "old", BackupName: "daily", StartTime: 100, State: "Backup_Success"},
		{BackupId: "listed", BackupName: "daily", StartTime: 500, State: "Backup_Running"},
	}
	if got := matchBackup(backups, "created", "daily", 400); got == nil || got.BackupId != "listed" {
		t.Fatalf("name match = %#v", got)
	}
	if got := matchBackup(backups, "created", "daily", 600); got != nil {
		t.Fatalf("old backup matched a newer create: %#v", got)
	}
	ambiguous := append(backups, umongodb.BackupInfo{BackupId: "other", BackupName: "daily", StartTime: 500})
	if got := matchBackup(ambiguous, "created", "daily", 400); got != nil {
		t.Fatalf("ambiguous backup matched: %#v", got)
	}
	if got := matchListedBackup(backups, "daily", 500); got == nil || got.BackupId != "listed" {
		t.Fatalf("listed backup = %#v", got)
	}
	if got := matchListedBackup(ambiguous, "daily", 500); got != nil {
		t.Fatalf("ambiguous listed backup matched: %#v", got)
	}
	if !shouldRestart("a", "b", statusRunning, false) || shouldRestart("a", "b", statusStopped, false) || shouldRestart("", "b", statusRunning, false) {
		t.Fatal("restart decision did not follow running status and a stored previous value")
	}
	if !clusterDeleteSettled(&umongodb.ClusterInfo{
		State:            "InitFailed",
		DataReplicaInfos: []umongodb.ReplicaInfo{{ReplicaId: "replica-1", State: apiStateDeleted}},
	}) {
		t.Fatal("deleted replicas on an InitFailed cluster were not settled")
	}
	if clusterDeleteSettled(&umongodb.ClusterInfo{
		State:            "InitFailed",
		DataReplicaInfos: []umongodb.ReplicaInfo{{ReplicaId: "replica-1", State: apiStateDeleted}},
		MongosInfo:       []umongodb.NodeInfo{{NodeId: "mongos-1", State: apiStateRunning}},
	}) {
		t.Fatal("live mongos was treated as deleted")
	}
	if clusterDeleteSettled(&umongodb.ClusterInfo{State: "InitFailed"}) {
		t.Fatal("InitFailed cluster without replicas was treated as deleted")
	}
}
