package uk8s

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
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
	fields["isolation_group"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
	}
	fields["name_prefix"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
	}
	fields["security_mode"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
		ValidateFunc: validation.StringInSlice([]string{"Firewall", "SecGroup"}, false),
	}
	fields["uni_feature"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
		ValidateFunc: validation.StringInSlice([]string{"true", "false"}, false),
	}
	fields["data_disk_kms_key_id"] = &schema.Schema{
		Type: schema.TypeString, Optional: true, ForceNew: true,
	}
	fields["security_group"] = &schema.Schema{
		Type: schema.TypeList, Optional: true, ForceNew: true, MaxItems: 5,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"id": {
				Type: schema.TypeString, Required: true, ForceNew: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`\S`), "must not be empty"),
			},
			"priority": {
				Type: schema.TypeInt, Optional: true, ForceNew: true,
				ValidateFunc: validation.IntBetween(1, 5),
			},
			"name": {Type: schema.TypeString, Optional: true, ForceNew: true},
		}},
	}
	fields["network_interface"] = &schema.Schema{
		Type: schema.TypeList, Optional: true, ForceNew: true,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"eip": {
				Type: schema.TypeList, Required: true, ForceNew: true, MinItems: 1, MaxItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"bandwidth": {
						Type: schema.TypeInt, Optional: true, ForceNew: true,
						ValidateFunc: validation.IntAtLeast(0),
					},
					"pay_mode": {
						Type: schema.TypeString, Optional: true, ForceNew: true, Default: "Bandwidth",
						ValidateFunc: validation.StringInSlice(
							[]string{"Traffic", "Bandwidth", "ShareBandwidth", "Free"}, false,
						),
					},
					"operator_name": {
						Type: schema.TypeString, Required: true, ForceNew: true,
						ValidateFunc: validation.StringInSlice([]string{"Bgp", "International"}, false),
					},
					"share_bandwidth_id": &schema.Schema{Type: schema.TypeString, Optional: true, ForceNew: true},
					"coupon_id":          &schema.Schema{Type: schema.TypeString, Optional: true, ForceNew: true},
				}},
			},
		}},
	}
	fields["kubelet_configuration"] = &schema.Schema{
		Type: schema.TypeMap, Optional: true, ForceNew: true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Kubelet configuration map supporting only ContainerLogMaxFiles, an integer string of at least 2.",
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

func expandUK8SWorkers(items []interface{}) ([]sdkuk8s.CreateUK8SClusterV2ParamNodes, error) {
	workers := make([]sdkuk8s.CreateUK8SClusterV2ParamNodes, 0, len(items))
	for i, item := range items {
		worker, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("worker.%d configuration is not available", i)
		}
		if strings.HasPrefix(worker["machine_type"].(string), "O") && worker["boot_disk_type"] != "cloud_rssd" {
			return nil, fmt.Errorf("worker.%d.boot_disk_type must be cloud_rssd for outstanding machines", i)
		}
		intelFamily := worker["uhost_family"] == "o1i" || worker["uhost_family"] == "o2i"
		intelPlatform := strings.HasPrefix(worker["min_cpu_platform"].(string), "Intel/")
		if intelFamily && (worker["machine_type"] != "O" || !intelPlatform) {
			return nil, fmt.Errorf("worker.%d.uhost_family requires machine_type O and an Intel CPU platform", i)
		}
		if len(worker["labels"].(map[string]interface{})) > 5 {
			return nil, fmt.Errorf("worker.%d.labels supports at most 5 entries", i)
		}
		labels, taints, err := expandUK8SNodeGroupScheduling(uk8sWorkerValues(worker))
		if err != nil {
			return nil, fmt.Errorf("worker.%d: %w", i, err)
		}
		node := sdkuk8s.CreateUK8SClusterV2ParamNodes{
			Zone:               ucloud.String(worker["availability_zone"].(string)),
			MachineType:        ucloud.String(worker["machine_type"].(string)),
			CPU:                ucloud.Int(worker["cpu"].(int)),
			Mem:                ucloud.Int(worker["memory"].(int)),
			Count:              ucloud.Int(worker["count"].(int)),
			GPU:                ucloud.Int(worker["gpu"].(int)),
			MaxPods:            ucloud.Int(worker["max_pods"].(int)),
			BootDiskType:       ucloud.String(upperCvt.unconvert(worker["boot_disk_type"].(string))),
			BootDiskSize:       ucloud.Int(worker["boot_disk_size"].(int)),
			MinimalCpuPlatform: ucloud.String(worker["min_cpu_platform"].(string)),
		}
		if worker["data_disk_size"].(int) > 0 {
			node.DataDiskSize = ucloud.Int(worker["data_disk_size"].(int))
			node.DataDiskType = ucloud.String(upperCvt.unconvert(worker["data_disk_type"].(string)))
		}
		if value := worker["image_id"].(string); value != "" {
			node.ImageId = ucloud.String(value)
		}
		if value := worker["uhost_family"].(string); value != "" {
			node.UHostFamily = ucloud.String(value)
		}
		if value := worker["security_group_id"].(string); value != "" {
			node.SecurityGroupId = ucloud.String(value)
		}
		if value := worker["gpu_type"].(string); value != "" {
			node.GpuType = ucloud.String(value)
		}
		if labels != "" {
			node.Labels = ucloud.String(labels)
		}
		if taints != "" {
			node.Taints = ucloud.String(taints)
		}
		if value := worker["isolation_group"].(string); value != "" {
			node.IsolationGroup = ucloud.String(value)
		}
		if value := worker["name_prefix"].(string); value != "" {
			node.NamePrefix = ucloud.String(value)
		}
		if value := worker["security_mode"].(string); value != "" {
			node.SecurityMode = ucloud.String(value)
		}
		if value := worker["uni_feature"].(string); value != "" {
			node.UNIFeature = ucloud.String(value)
		}
		if value := worker["data_disk_kms_key_id"].(string); value != "" {
			node.DataDiskKmsKeyId = ucloud.String(value)
		}
		groups := worker["security_group"].([]interface{})
		if err := validateUK8SSecurityGroups(groups); err != nil {
			return nil, fmt.Errorf("worker.%d: %w", i, err)
		}
		if len(groups) > 0 && worker["security_mode"] != "SecGroup" {
			return nil, fmt.Errorf("worker.%d: security_group requires security_mode SecGroup", i)
		}
		if worker["security_mode"] == "SecGroup" && worker["security_group_id"] != "" {
			return nil, fmt.Errorf("worker.%d: security_group_id is a firewall ID and cannot be used with security_mode SecGroup", i)
		}
		for _, item := range groups {
			group := item.(map[string]interface{})
			binding := sdkuk8s.CreateUK8SClusterV2ParamNodesSecGroupId{Id: ucloud.String(group["id"].(string))}
			if priority := group["priority"].(int); priority != 0 {
				binding.Priority = ucloud.String(strconv.Itoa(priority))
			}
			if name := group["name"].(string); name != "" {
				binding.Name = ucloud.String(name)
			}
			node.SecGroupId = append(node.SecGroupId, binding)
		}
		for interfaceIndex, item := range worker["network_interface"].([]interface{}) {
			config := item.(map[string]interface{})["eip"].([]interface{})[0].(map[string]interface{})
			eip, err := expandUK8SEIP(config)
			if err != nil {
				return nil, fmt.Errorf("worker.%d: network_interface.%d.eip: %w", i, interfaceIndex, err)
			}
			node.NetworkInterface = append(node.NetworkInterface, sdkuk8s.CreateUK8SClusterV2ParamNodesNetworkInterface{EIP: eip})
		}
		kubelet, err := expandUK8SKubeletConfiguration(worker["kubelet_configuration"].(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("worker.%d: %w", i, err)
		}
		node.KubeletConfiguration = kubelet
		workers = append(workers, node)
	}
	return workers, nil
}

func expandUK8SEIP(config map[string]interface{}) (*sdkuk8s.CreateUK8SClusterV2ParamNodesNetworkInterfaceEIP, error) {
	mode := config["pay_mode"].(string)
	bandwidth := config["bandwidth"].(int)
	sharedID := config["share_bandwidth_id"].(string)
	if mode == "ShareBandwidth" {
		if sharedID == "" {
			return nil, fmt.Errorf("ShareBandwidth requires share_bandwidth_id")
		}
	} else {
		if bandwidth <= 0 {
			return nil, fmt.Errorf("bandwidth must be positive outside ShareBandwidth mode")
		}
		if sharedID != "" {
			return nil, fmt.Errorf("share_bandwidth_id requires pay_mode ShareBandwidth")
		}
	}
	if mode == "Traffic" && bandwidth > 300 {
		return nil, fmt.Errorf("Traffic bandwidth must be between 1 and 300 Mbps")
	}
	if mode == "Bandwidth" && bandwidth > 800 {
		return nil, fmt.Errorf("Bandwidth bandwidth must be between 1 and 800 Mbps")
	}
	eip := &sdkuk8s.CreateUK8SClusterV2ParamNodesNetworkInterfaceEIP{
		PayMode:      ucloud.String(mode),
		OperatorName: ucloud.String(config["operator_name"].(string)),
	}
	if bandwidth > 0 {
		eip.Bandwidth = ucloud.Int(bandwidth)
	}
	if sharedID != "" {
		eip.ShareBandwidthId = ucloud.String(sharedID)
	}
	if coupon := config["coupon_id"].(string); coupon != "" {
		eip.CouponId = ucloud.String(coupon)
	}
	return eip, nil
}

func expandUK8SKubeletConfiguration(
	fields map[string]interface{},
) (*sdkuk8s.CreateUK8SClusterV2ParamNodesKubeletConfiguration, error) {
	if len(fields) == 0 {
		return nil, nil
	}
	for key := range fields {
		if key != "ContainerLogMaxFiles" {
			return nil, fmt.Errorf("kubelet_configuration only supports ContainerLogMaxFiles, got %q", key)
		}
	}
	value := fields["ContainerLogMaxFiles"].(string)
	count, err := strconv.Atoi(value)
	if err != nil || count < 2 {
		return nil, fmt.Errorf("kubelet_configuration.ContainerLogMaxFiles must be an integer string of at least 2")
	}
	return &sdkuk8s.CreateUK8SClusterV2ParamNodesKubeletConfiguration{
		ContainerLogMaxFiles: ucloud.String(value),
	}, nil
}
