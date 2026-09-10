package uk8s

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

const (
	maxUK8SNodeGroupLabels = 20
	maxUK8SNodeGroupTaints = 10
)

func resourceUCloudUK8SNodeGroup() *schema.Resource {
	return &schema.Resource{
		Create:   resourceUK8SNodeGroupCreate,
		Read:     resourceUK8SNodeGroupRead,
		Update:   resourceUK8SNodeGroupUpdate,
		Delete:   resourceUK8SNodeGroupDelete,
		Importer: &schema.ResourceImporter{State: importUK8SNodeGroup},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(diffValidateUK8SNodeGroupScheduling),

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type: schema.TypeString, Required: true, ForceNew: true,
			},
			"name": {
				Type: schema.TypeString, Required: true, ValidateFunc: validateName,
			},
			"availability_zone": {
				Type: schema.TypeString, Required: true,
			},
			"subnet_id": {
				Type: schema.TypeString, Required: true,
			},
			"image_id": {
				Type: schema.TypeString, Optional: true,
			},
			"instance_type": {
				Type: schema.TypeString, Required: true, ValidateFunc: validateInstanceType,
			},
			"uhost_family": {
				Type: schema.TypeString, Optional: true, Computed: true,
				ValidateFunc: validation.StringInSlice([]string{"o1i", "o2i"}, false),
			},
			"charge_type": {
				Type: schema.TypeString, Optional: true, Default: "month",
				ValidateFunc: validation.StringInSlice([]string{"year", "month", "dynamic"}, false),
			},
			"boot_disk_type": {
				Type: schema.TypeString, Optional: true, Default: "cloud_ssd",
				ValidateFunc: validation.StringInSlice([]string{
					"local_normal", "local_ssd", "cloud_normal", "cloud_ssd", "cloud_rssd",
				}, false),
			},
			"boot_disk_size": {
				Type: schema.TypeInt, Optional: true, Default: 40,
				ValidateFunc: validation.IntBetween(40, 500),
			},
			"data_disk_type": {
				Type: schema.TypeString, Optional: true, Default: "cloud_ssd",
				ValidateFunc: validation.StringInSlice([]string{
					"local_normal", "local_ssd", "cloud_normal", "cloud_ssd", "cloud_rssd",
				}, false),
			},
			"data_disk_size": {
				Type: schema.TypeInt, Optional: true, Default: 0,
				ValidateFunc: validateAll(validation.IntBetween(0, 2000), validateMod(10)),
			},
			"isolation_group": {
				Type: schema.TypeString, Optional: true,
			},
			"min_cpu_platform": {
				Type: schema.TypeString, Optional: true, Default: "Intel/Auto",
				ValidateFunc: validation.StringInSlice([]string{
					"Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell",
					"Intel/Skylake", "Intel/Cascadelake", "Intel/CascadelakeR",
					"Amd/Auto", "Amd/Epyc2", "Ampere/Altra",
				}, false),
			},
			"max_pods": {
				Type: schema.TypeInt, Optional: true, Default: 110,
				ValidateFunc: validation.IntBetween(1, 256),
			},
			"tag": {
				Type: schema.TypeString, Optional: true,
			},
			"user_data": {
				Type: schema.TypeString, Optional: true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"init_script": {
				Type: schema.TypeString, Optional: true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"labels": {
				Type: schema.TypeMap, Optional: true,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"taint": {
				Type: schema.TypeSet, Optional: true, MaxItems: maxUK8SNodeGroupTaints,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":    {Type: schema.TypeString, Required: true},
					"value":  {Type: schema.TypeString, Optional: true},
					"effect": {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringInSlice([]string{"NoExecute", "NoSchedule", "PreferNoSchedule"}, false)},
				}},
			},
			"node_ids": {
				Type: schema.TypeSet, Computed: true, Elem: &schema.Schema{Type: schema.TypeString},
			},
			"create_time": {Type: schema.TypeString, Computed: true},
			"update_time": {Type: schema.TypeString, Computed: true},
		},
	}
}

func importUK8SNodeGroup(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	parts := strings.Split(importID, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("invalid uk8s node group import ID %q: expected <cluster_id>/<node_group_id>", importID)
	}
	if err := d.Set("cluster_id", parts[0]); err != nil {
		return nil, fmt.Errorf("import uk8s node group %q: %w", importID, err)
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

func resourceUK8SNodeGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating uk8s node group, %s", err)
	}
	req, err := buildUK8SNodeGroupAddRequest(client, d)
	if err != nil {
		return err
	}
	resp, err := client.AddUK8SNodeGroup(req)
	if err != nil {
		return fmt.Errorf("error on creating uk8s node group, %s", err)
	}
	if resp.NodeGroupId == "" {
		return fmt.Errorf("error on creating uk8s node group, response contains no node group id")
	}
	d.SetId(resp.NodeGroupId)
	return resourceUK8SNodeGroupRead(d, meta)
}

func resourceUK8SNodeGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating uk8s node group, %s", err)
	}
	req, err := buildUK8SNodeGroupUpdateRequest(client, d, d.Id())
	if err != nil {
		return err
	}
	if _, err := client.UpdateUK8SNodeGroup(req); err != nil {
		return fmt.Errorf("error on updating uk8s node group %q, %s", d.Id(), err)
	}
	return resourceUK8SNodeGroupRead(d, meta)
}

func resourceUK8SNodeGroupRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading uk8s node group, %s", err)
	}
	groups, err := listUK8SNodeGroups(client, d.Get("cluster_id").(string))
	if err != nil {
		return fmt.Errorf("error on reading uk8s node group %q, %s", d.Id(), err)
	}
	for _, group := range groups {
		if group.NodeGroupId != d.Id() {
			continue
		}
		setUK8SNodeGroupState(d, group)
		return nil
	}
	d.SetId("")
	return nil
}

func listUK8SNodeGroups(client *sdkuk8s.UK8SClient, clusterID string) ([]sdkuk8s.NodeGroupSet, error) {
	req := client.NewListUK8SNodeGroupRequest()
	req.ClusterId = ucloud.String(clusterID)
	resp, err := client.ListUK8SNodeGroup(req)
	if err != nil {
		return nil, err
	}
	return resp.NodeGroupList, nil
}

func resourceUK8SNodeGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting uk8s node group %q: %w", d.Id(), err)
	}
	req := client.NewRemoveUK8SNodeGroupRequest()
	req.ClusterId = ucloud.String(d.Get("cluster_id").(string))
	req.NodeGroupId = ucloud.String(d.Id())
	if _, err := client.RemoveUK8SNodeGroup(req); err != nil {
		return fmt.Errorf("error on deleting uk8s node group %q, %s", d.Id(), err)
	}
	d.SetId("")
	return nil
}

func buildUK8SNodeGroupAddRequest(client *sdkuk8s.UK8SClient, d *schema.ResourceData) (*sdkuk8s.AddUK8SNodeGroupRequest, error) {
	instance, err := parseInstanceType(d.Get("instance_type").(string))
	if err != nil {
		return nil, err
	}
	labels, taints, err := expandUK8SNodeGroupScheduling(d)
	if err != nil {
		return nil, err
	}
	req := client.NewAddUK8SNodeGroupRequest()
	req.ClusterId = ucloud.String(d.Get("cluster_id").(string))
	req.NodeGroupName = ucloud.String(d.Get("name").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	req.MachineType = ucloud.String(strings.ToUpper(instance.HostType))
	req.MinimalCpuPlatform = ucloud.String(d.Get("min_cpu_platform").(string))
	req.MaxPods = ucloud.String(strconv.Itoa(d.Get("max_pods").(int)))
	req.CPU = ucloud.Int(instance.CPU)
	req.Mem = ucloud.Int(instance.Memory)
	req.Labels = ucloud.String(labels)
	req.Taints = ucloud.String(taints)
	req.BootDiskType = ucloud.String(upperCvt.unconvert(d.Get("boot_disk_type").(string)))
	req.BootDiskSize = ucloud.Int(d.Get("boot_disk_size").(int))
	req.DataDiskSize = ucloud.Int(d.Get("data_disk_size").(int))
	req.DataDiskType = ucloud.String(upperCvt.unconvert(d.Get("data_disk_type").(string)))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	if value, ok := d.GetOk("uhost_family"); ok {
		req.UHostFamily = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("image_id"); ok {
		req.ImageId = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("isolation_group"); ok {
		req.IsolationGroupId = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("user_data"); ok {
		req.UserData = ucloud.String(base64.StdEncoding.EncodeToString([]byte(value.(string))))
	}
	if value, ok := d.GetOk("init_script"); ok {
		req.InitScript = ucloud.String(base64.StdEncoding.EncodeToString([]byte(value.(string))))
	}
	return req, nil
}

func buildUK8SNodeGroupUpdateRequest(client *sdkuk8s.UK8SClient, d *schema.ResourceData, nodeGroupID string) (*sdkuk8s.UpdateUK8SNodeGroupRequest, error) {
	addReq, err := buildUK8SNodeGroupAddRequest(client, d)
	if err != nil {
		return nil, err
	}
	req := client.NewUpdateUK8SNodeGroupRequest()
	req.Zone = addReq.Zone
	req.ClusterId = addReq.ClusterId
	req.NodeGroupId = ucloud.String(nodeGroupID)
	req.NodeGroupName = addReq.NodeGroupName
	req.ImageId = addReq.ImageId
	req.SubnetId = addReq.SubnetId
	req.MachineType = addReq.MachineType
	req.UHostFamily = addReq.UHostFamily
	req.MinimalCpuPlatform = addReq.MinimalCpuPlatform
	req.IsolationGroupId = addReq.IsolationGroupId
	req.MaxPods = addReq.MaxPods
	req.CPU = addReq.CPU
	req.Mem = addReq.Mem
	req.Labels = addReq.Labels
	req.Taints = addReq.Taints
	req.BootDiskType = addReq.BootDiskType
	req.BootDiskSize = addReq.BootDiskSize
	req.DataDiskSize = addReq.DataDiskSize
	req.DataDiskType = addReq.DataDiskType
	req.Tag = addReq.Tag
	req.UserData = addReq.UserData
	req.InitScript = addReq.InitScript
	req.ChargeType = addReq.ChargeType
	return req, nil
}

func setUK8SNodeGroupState(d *schema.ResourceData, group sdkuk8s.NodeGroupSet) {
	_ = d.Set("name", group.NodeGroupName)
	_ = d.Set("availability_zone", group.Zone)
	_ = d.Set("subnet_id", group.SubnetId)
	_ = d.Set("image_id", group.ImageId)
	_ = d.Set("uhost_family", group.UHostFamily)
	_ = d.Set("instance_type", formatUK8SInstanceType(group.MachineType, group.CPU, group.Mem))
	_ = d.Set("charge_type", upperCamelCvt.convert(group.ChargeType))
	_ = d.Set("boot_disk_type", upperCvt.convert(group.BootDiskType))
	_ = d.Set("boot_disk_size", group.BootDiskSize)
	_ = d.Set("data_disk_size", group.DataDiskSize)
	dataDiskType := upperCvt.convert(group.DataDiskType)
	if dataDiskType == "" {
		dataDiskType = "cloud_ssd"
	}
	_ = d.Set("data_disk_type", dataDiskType)
	_ = d.Set("isolation_group", group.IsolationGroupId)
	_ = d.Set("min_cpu_platform", group.MinimalCpuPlatform)
	_ = d.Set("max_pods", group.MaxPods)
	_ = d.Set("tag", group.Tag)
	_ = d.Set("user_data", decodeUK8SNodeGroupScript(group.UserData))
	_ = d.Set("init_script", decodeUK8SNodeGroupScript(group.InitScript))
	_ = d.Set("labels", flattenUK8SNodeGroupLabels(group.Labels))
	_ = d.Set("taint", flattenUK8SNodeGroupTaints(group.Taints))
	_ = d.Set("node_ids", group.NodeList)
	if group.CreateTime > 0 {
		_ = d.Set("create_time", timestampToString(group.CreateTime))
	}
	if group.UpdateTime > 0 {
		_ = d.Set("update_time", timestampToString(group.UpdateTime))
	}
}

func decodeUK8SNodeGroupScript(value string) string {
	if value == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return value
	}
	return string(decoded)
}

func formatUK8SInstanceType(machineType string, cpu, memory int) string {
	hostType := strings.ToLower(machineType)
	for mode, scale := range instanceTypeScaleMap {
		if cpu > 0 && memory == cpu*scale {
			return fmt.Sprintf("%s-%s-%d", hostType, mode, cpu)
		}
	}
	return fmt.Sprintf("%s-customized-%d-%d", hostType, cpu, memory/1024)
}

func diffValidateUK8SNodeGroupScheduling(diff *schema.ResourceDiff, meta interface{}) error {
	_, _, err := expandUK8SNodeGroupScheduling(diff)
	return err
}

type resourceValueGetter interface {
	Get(string) interface{}
}

func expandUK8SNodeGroupScheduling(d resourceValueGetter) (string, string, error) {
	rawLabels := d.Get("labels").(map[string]interface{})
	if len(rawLabels) > maxUK8SNodeGroupLabels {
		return "", "", fmt.Errorf("labels supports at most %d entries", maxUK8SNodeGroupLabels)
	}
	labels := make([]string, 0, len(rawLabels))
	for key, rawValue := range rawLabels {
		value := rawValue.(string)
		if err := validateUK8SSchedulingKeyValue("label", key, value); err != nil {
			return "", "", err
		}
		labels = append(labels, key+"="+value)
	}
	sort.Strings(labels)

	rawTaints := d.Get("taint").(*schema.Set).List()
	taints := make([]string, 0, len(rawTaints))
	keys := make(map[string]struct{}, len(rawTaints))
	for _, raw := range rawTaints {
		item := raw.(map[string]interface{})
		key := item["key"].(string)
		value := item["value"].(string)
		if _, exists := keys[key]; exists {
			return "", "", fmt.Errorf("taint key %q is duplicated", key)
		}
		keys[key] = struct{}{}
		if err := validateUK8SSchedulingKeyValue("taint", key, value); err != nil {
			return "", "", err
		}
		taints = append(taints, fmt.Sprintf("%s=%s:%s", key, value, item["effect"].(string)))
	}
	sort.Strings(taints)
	return strings.Join(labels, ","), strings.Join(taints, ","), nil
}

func validateUK8SSchedulingKeyValue(kind, key, value string) error {
	if key == "" || strings.ContainsAny(key, "=,") {
		return fmt.Errorf("%s key %q is invalid", kind, key)
	}
	prefix := strings.SplitN(key, "/", 2)[0]
	if strings.Contains(prefix, "kubernetes.io") || strings.Contains(prefix, "k8s.io") {
		return fmt.Errorf("%s key %q uses a reserved Kubernetes prefix", kind, key)
	}
	if strings.ContainsAny(value, "=,:") {
		return fmt.Errorf("%s value %q contains an unsupported delimiter", kind, value)
	}
	return nil
}

func flattenUK8SNodeGroupLabels(value string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, item := range strings.Split(value, ",") {
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func flattenUK8SNodeGroupTaints(value string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, item := range strings.Split(value, ",") {
		if item == "" {
			continue
		}
		keyValue := strings.SplitN(item, "=", 2)
		if len(keyValue) != 2 {
			continue
		}
		valueEffect := strings.SplitN(keyValue[1], ":", 2)
		if len(valueEffect) != 2 {
			continue
		}
		result = append(result, map[string]interface{}{
			"key": keyValue[0], "value": valueEffect[0], "effect": valueEffect[1],
		})
	}
	return result
}
