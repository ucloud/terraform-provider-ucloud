package umongodb

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
)

const (
	kindReplica = "replicaset"
	kindSharded = "sharded"

	clusterTypeReplica = "ReplicaSet"
	clusterTypeSharded = "SharedCluster"

	statusRunning = "running"
	statusStopped = "stopped"

	apiStateRunning = "Running"
	apiStateStopped = "Stopped"
	apiStateDeleted = "Deleted"
)

func clusterTimeouts() *schema.ResourceTimeout {
	return &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(60 * time.Minute),
		Update: schema.DefaultTimeout(45 * time.Minute),
		Delete: schema.DefaultTimeout(30 * time.Minute),
	}
}

func clusterImporter() *schema.ResourceImporter {
	return &schema.ResourceImporter{State: schema.ImportStatePassthrough}
}

func baseClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"availability_zone": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(6, 63),
		},
		"db_version": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"admin_password": {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		"disk_space": {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validateDiskSpace,
		},
		"machine_type_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"node_count": {
			Type:         schema.TypeInt,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IntInSlice([]int{3, 5, 7}),
		},
		"listen_port": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
		},
		"subnet_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"vpc_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"tag": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"charge_type": {
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			Default:          "month",
			DiffSuppressFunc: suppressUnreadableDefault("month"),
			ValidateFunc: validation.StringInSlice([]string{
				"year", "month", "dynamic", "trial",
			}, false),
		},
		"quantity": {
			Type:             schema.TypeInt,
			Optional:         true,
			ForceNew:         true,
			Default:          1,
			DiffSuppressFunc: suppressUnreadableDefault("1"),
			ValidateFunc:     validation.IntAtLeast(1),
		},
		"cross_zones": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"label": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"key":   {Type: schema.TypeString, Required: true},
				"value": {Type: schema.TypeString, Required: true},
			}},
		},
		"sec_group": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"id":       {Type: schema.TypeString, Required: true},
				"priority": {Type: schema.TypeInt, Required: true, ValidateFunc: validation.IntBetween(1, 5)},
			}},
		},
		"status": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      statusRunning,
			ValidateFunc: validation.StringInSlice([]string{statusRunning, statusStopped}, false),
		},
		"restart_trigger": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"auto_backup_week": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"auto_backup_time": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"backup_disabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"state": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"connect_url": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"create_time": {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}
}

func shardedClusterSchema() map[string]*schema.Schema {
	fields := baseClusterSchema()
	fields["shard_count"] = &schema.Schema{
		Type:         schema.TypeInt,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.IntBetween(2, 9),
	}
	fields["mongos_node_count"] = &schema.Schema{
		Type:         schema.TypeInt,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.IntBetween(2, 5),
	}
	fields["mongos_machine_type_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}
	// CreateUMongoDBShardedCluster has no CrossZones field.
	delete(fields, "cross_zones")
	return fields
}

func validateDiskSpace(value interface{}, key string) ([]string, []error) {
	space, ok := value.(int)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %s to be int", key)}
	}
	if space < 20 || space > 32000 || space%10 != 0 {
		return nil, []error{fmt.Errorf("%s must be a multiple of 10 between 20 and 32000", key)}
	}
	return nil, nil
}

func chargeTypeAPI(value string) string {
	switch value {
	case "year":
		return "Year"
	case "month":
		return "Month"
	case "dynamic":
		return "Dynamic"
	case "trial":
		return "Trial"
	default:
		return value
	}
}

func suppressUnreadableDefault(defaultValue string) schema.SchemaDiffSuppressFunc {
	return func(key, old, new string, data *schema.ResourceData) bool {
		return old == "" && new == defaultValue
	}
}

func expandStringList(raw interface{}) []string {
	if set, ok := raw.(*schema.Set); ok {
		raw = set.List()
	}
	items, ok := raw.([]interface{})
	if !ok || len(items) == 0 {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok && text != "" {
			values = append(values, text)
		}
	}
	return values
}

func expandReplSetLabels(raw interface{}) []umongodb.CreateUMongoDBReplSetParamLabels {
	set, ok := raw.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil
	}
	labels := make([]umongodb.CreateUMongoDBReplSetParamLabels, 0, set.Len())
	for _, item := range set.List() {
		row := item.(map[string]interface{})
		labels = append(labels, umongodb.CreateUMongoDBReplSetParamLabels{
			Key:   stringPtr(row["key"].(string)),
			Value: stringPtr(row["value"].(string)),
		})
	}
	return labels
}

func expandShardedLabels(raw interface{}) []umongodb.CreateUMongoDBShardedClusterParamLabels {
	set, ok := raw.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil
	}
	labels := make([]umongodb.CreateUMongoDBShardedClusterParamLabels, 0, set.Len())
	for _, item := range set.List() {
		row := item.(map[string]interface{})
		labels = append(labels, umongodb.CreateUMongoDBShardedClusterParamLabels{
			Key:   stringPtr(row["key"].(string)),
			Value: stringPtr(row["value"].(string)),
		})
	}
	return labels
}

func expandReplSetSecGroups(raw interface{}) []umongodb.CreateUMongoDBReplSetParamSecGroupId {
	set, ok := raw.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil
	}
	groups := make([]umongodb.CreateUMongoDBReplSetParamSecGroupId, 0, set.Len())
	for _, item := range set.List() {
		row := item.(map[string]interface{})
		groups = append(groups, umongodb.CreateUMongoDBReplSetParamSecGroupId{
			Id:       stringPtr(row["id"].(string)),
			Priority: intPtr(row["priority"].(int)),
		})
	}
	return groups
}

func expandShardedSecGroups(raw interface{}) []umongodb.CreateUMongoDBShardedClusterParamSecGroupId {
	set, ok := raw.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil
	}
	groups := make([]umongodb.CreateUMongoDBShardedClusterParamSecGroupId, 0, set.Len())
	for _, item := range set.List() {
		row := item.(map[string]interface{})
		groups = append(groups, umongodb.CreateUMongoDBShardedClusterParamSecGroupId{
			Id:       stringPtr(row["id"].(string)),
			Priority: intPtr(row["priority"].(int)),
		})
	}
	return groups
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func replSetNodeCount(info *umongodb.ClusterInfo) int {
	for _, replica := range info.DataReplicaInfos {
		if replica.ReplicaType == "" || replica.ReplicaType == "DataRepl" {
			if replica.NodeCount > 0 {
				return replica.NodeCount
			}
			if len(replica.NodeInfos) > 0 {
				return len(replica.NodeInfos)
			}
		}
	}
	return 0
}
