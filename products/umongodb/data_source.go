package umongodb

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
)

func dataSourceUCloudUMongoDBInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUMongoDBInstanceRead,
		Schema: map[string]*schema.Schema{
			"availability_zone": {Type: schema.TypeString, Optional: true, Computed: true},
			"cluster_id":        {Type: schema.TypeString, Required: true},
			"cluster_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					clusterTypeReplica, clusterTypeSharded,
				}, false),
			},
			"name":            {Type: schema.TypeString, Computed: true},
			"db_version":      {Type: schema.TypeString, Computed: true},
			"disk_space":      {Type: schema.TypeInt, Computed: true},
			"machine_type_id": {Type: schema.TypeString, Computed: true},
			"node_count":      {Type: schema.TypeInt, Computed: true},
			"shard_count":     {Type: schema.TypeInt, Computed: true},
			"mongos_count":    {Type: schema.TypeInt, Computed: true},
			"state":           {Type: schema.TypeString, Computed: true},
			"connect_url":     {Type: schema.TypeString, Computed: true},
			"subnet_id":       {Type: schema.TypeString, Computed: true},
			"vpc_id":          {Type: schema.TypeString, Computed: true},
			"tag":             {Type: schema.TypeString, Computed: true},
			"create_time":     {Type: schema.TypeInt, Computed: true},
		},
	}
}

func dataSourceUCloudUMongoDBInstanceRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading umongodb instance, %s", err)
	}
	clusterID := data.Get("cluster_id").(string)
	info, err := describeCluster(client, clusterID, data.Get("cluster_type").(string), data.Get("availability_zone").(string))
	if err != nil {
		return fmt.Errorf("error on reading umongodb instance %q, %s", clusterID, err)
	}
	data.SetId(info.ClusterId)
	_ = data.Set("name", info.InstanceName)
	_ = data.Set("db_version", info.DBVersion)
	_ = data.Set("disk_space", info.DiskSpace)
	_ = data.Set("machine_type_id", info.MachineTypeId)
	_ = data.Set("state", info.State)
	_ = data.Set("connect_url", info.ConnectURL)
	_ = data.Set("subnet_id", info.SubnetId)
	_ = data.Set("vpc_id", info.VPCId)
	_ = data.Set("tag", info.Tag)
	_ = data.Set("create_time", info.CreateTime)
	_ = data.Set("availability_zone", info.Zone)
	_ = data.Set("cluster_type", info.ClusterType)
	_ = data.Set("shard_count", info.ShardCount)
	_ = data.Set("mongos_count", info.MongosCount)
	if info.ClusterType == clusterTypeSharded && info.ShardNodeCount > 0 {
		_ = data.Set("node_count", info.ShardNodeCount)
	} else if count := replSetNodeCount(info); count > 0 {
		_ = data.Set("node_count", count)
	}
	return nil
}

func dataSourceUCloudUMongoDBInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUMongoDBInstancesRead,
		Schema: map[string]*schema.Schema{
			"availability_zone": {Type: schema.TypeString, Optional: true},
			"cluster_id":        {Type: schema.TypeString, Optional: true},
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":           {Type: schema.TypeString, Computed: true},
					"name":         {Type: schema.TypeString, Computed: true},
					"cluster_type": {Type: schema.TypeString, Computed: true},
					"db_version":   {Type: schema.TypeString, Computed: true},
					"disk_space":   {Type: schema.TypeInt, Computed: true},
					"state":        {Type: schema.TypeString, Computed: true},
					"connect_url":  {Type: schema.TypeString, Computed: true},
					"zone":         {Type: schema.TypeString, Computed: true},
					"subnet_id":    {Type: schema.TypeString, Computed: true},
					"vpc_id":       {Type: schema.TypeString, Computed: true},
					"tag":          {Type: schema.TypeString, Computed: true},
					"create_time":  {Type: schema.TypeInt, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudUMongoDBInstancesRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when listing umongodb instances, %s", err)
	}
	clusterID := data.Get("cluster_id").(string)
	zone := data.Get("availability_zone").(string)
	instances, err := listInstances(client, clusterID, zone)
	if err != nil {
		return err
	}
	instances = filterInstances(instances, clusterID, zone)
	clusters := make([]map[string]interface{}, 0, len(instances))
	ids := make([]string, 0, len(instances))
	for _, instance := range instances {
		clusters = append(clusters, flattenInstance(instance))
		ids = append(ids, instance.ClusterId)
	}
	if err := data.Set("clusters", clusters); err != nil {
		return err
	}
	data.SetId(dataSourceID(ids))
	return nil
}

func flattenInstance(instance umongodb.MongodbInstance) map[string]interface{} {
	return map[string]interface{}{
		"id":           instance.ClusterId,
		"name":         instance.Name,
		"cluster_type": instance.ClusterType,
		"db_version":   instance.DBVersion,
		"disk_space":   instance.DiskSpace,
		"state":        instance.State,
		"connect_url":  instance.ConnectURL,
		"zone":         instance.Zone,
		"subnet_id":    instance.SubnetId,
		"vpc_id":       instance.VPCId,
		"tag":          instance.Tag,
		"create_time":  instance.CreateTime,
	}
}

func dataSourceID(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

func filterInstances(instances []umongodb.MongodbInstance, clusterID, zone string) []umongodb.MongodbInstance {
	if clusterID == "" && zone == "" {
		return instances
	}
	filtered := make([]umongodb.MongodbInstance, 0, len(instances))
	for _, instance := range instances {
		if clusterID != "" && instance.ClusterId != clusterID {
			continue
		}
		if zone != "" && instance.Zone != zone {
			continue
		}
		filtered = append(filtered, instance)
	}
	return filtered
}
