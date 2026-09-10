package uk8s

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func uk8sClusterWorkerSchema(master *schema.Resource) *schema.Schema {
	fields := make(map[string]*schema.Schema)
	for _, key := range []string{
		"machine_type", "cpu", "memory", "uhost_family", "image_id",
		"boot_disk_type", "boot_disk_size", "data_disk_type", "data_disk_size", "min_cpu_platform",
	} {
		field := *master.Schema[key]
		field.Computed = false
		field.DiffSuppressFunc = nil
		fields[key] = &field
	}
	for _, key := range []string{"machine_type", "cpu", "memory"} {
		fields[key].Optional = false
		fields[key].Required = true
	}
	fields["boot_disk_type"].Default = "cloud_ssd"
	fields["boot_disk_size"].Default = 40
	fields["data_disk_type"].Default = "cloud_ssd"
	fields["data_disk_size"].Default = 0
	fields["data_disk_size"].ValidateFunc = func(value interface{}, key string) ([]string, []error) {
		size := value.(int)
		if size != 0 && (size < 20 || size > 1000 || size%10 != 0) {
			return nil, []error{fmt.Errorf("%s must be 0 or 20–1000 GB in multiples of 10", key)}
		}
		return nil, nil
	}
	fields["availability_zone"] = &schema.Schema{
		Type: schema.TypeString, Required: true, ForceNew: true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	fields["count"] = &schema.Schema{
		Type: schema.TypeInt, Optional: true, ForceNew: true, Default: 1,
		ValidateFunc: validation.IntBetween(1, 10),
	}
	fields["gpu"] = &schema.Schema{
		Type: schema.TypeInt, Optional: true, ForceNew: true, Default: 0,
		ValidateFunc: validation.IntAtLeast(0),
	}
	fields["gpu_type"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
	}
	fields["net_capability"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
		ValidateFunc: validation.StringInSlice([]string{"Normal", "Super", "Ultra", "Extreme"}, false),
	}
	fields["max_pods"] = &schema.Schema{
		Type: schema.TypeInt, Optional: true, ForceNew: true, Default: 110,
		ValidateFunc: validation.IntBetween(1, 256),
	}
	fields["security_group_id"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
	}
	fields["labels"] = &schema.Schema{
		Type: schema.TypeMap, Optional: true, ForceNew: true,
		Elem: &schema.Schema{Type: schema.TypeString},
	}
	fields["taint"] = &schema.Schema{
		Type: schema.TypeSet, Optional: true, ForceNew: true, MaxItems: 5,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"key":    {Type: schema.TypeString, Required: true},
			"value":  {Type: schema.TypeString, Optional: true},
			"effect": {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"NoExecute", "NoSchedule", "PreferNoSchedule"}, false)},
		}},
	}
	return &schema.Schema{
		Type: schema.TypeList, Optional: true, ForceNew: true,
		Description: "Initial Worker groups created with the cluster. Changes replace the entire cluster.",
		Elem:        &schema.Resource{Schema: fields},
	}
}

type uk8sWorkerValues map[string]interface{}

func (w uk8sWorkerValues) Get(key string) interface{} { return w[key] }

func diffValidateUK8SWorkers(diff *schema.ResourceDiff, meta interface{}) error {
	if !diff.NewValueKnown("worker") {
		return nil
	}
	// A known list can still contain attributes resolved only during apply.
	for _, key := range diff.GetChangedKeysPrefix("worker.") {
		if !diff.NewValueKnown(key) {
			return nil
		}
	}
	_, err := expandUK8SWorkers(diff.Get("worker").([]interface{}))
	return err
}

func expandUK8SWorkers(items []interface{}) ([]map[string]interface{}, error) {
	workers := make([]map[string]interface{}, 0, len(items))
	for i, item := range items {
		worker, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("worker.%d configuration is not available", i)
		}
		if strings.HasPrefix(worker["machine_type"].(string), "O") && worker["boot_disk_type"] != "cloud_rssd" {
			return nil, fmt.Errorf("worker.%d.boot_disk_type must be cloud_rssd for outstanding machines", i)
		}
		if worker["uhost_family"] != "" && (worker["machine_type"] != "O" || !strings.HasPrefix(worker["min_cpu_platform"].(string), "Intel/")) {
			return nil, fmt.Errorf("worker.%d.uhost_family requires machine_type O and an Intel CPU platform", i)
		}
		if len(worker["labels"].(map[string]interface{})) > 5 {
			return nil, fmt.Errorf("worker.%d.labels supports at most 5 entries", i)
		}
		labels, taints, err := expandUK8SNodeGroupScheduling(uk8sWorkerValues(worker))
		if err != nil {
			return nil, fmt.Errorf("worker.%d: %w", i, err)
		}
		// Use the wire names directly: the pinned SDK misspells BootDiskSize
		// and omits UHostFamily and NetCapability from its Nodes struct.
		node := map[string]interface{}{
			"Zone": worker["availability_zone"], "MachineType": worker["machine_type"],
			"CPU": worker["cpu"], "Mem": worker["memory"], "Count": worker["count"],
			"GPU": worker["gpu"], "MaxPods": worker["max_pods"],
			"BootDiskType":       upperCvt.unconvert(worker["boot_disk_type"].(string)),
			"BootDiskSize":       worker["boot_disk_size"],
			"MinimalCpuPlatform": worker["min_cpu_platform"],
		}
		if worker["data_disk_size"].(int) > 0 {
			node["DataDiskSize"] = worker["data_disk_size"]
			node["DataDiskType"] = upperCvt.unconvert(worker["data_disk_type"].(string))
		}
		for key, param := range map[string]string{
			"image_id": "ImageId", "uhost_family": "UHostFamily", "net_capability": "NetCapability",
			"security_group_id": "SecurityGroupId", "gpu_type": "GpuType",
		} {
			if value := worker[key].(string); value != "" {
				node[param] = value
			}
		}
		if labels != "" {
			node["Labels"] = labels
		}
		if taints != "" {
			node["Taints"] = taints
		}
		workers = append(workers, node)
	}
	return workers, nil
}
