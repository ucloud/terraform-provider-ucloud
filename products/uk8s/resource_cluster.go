package uk8s

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

func resourceUCloudUK8SCluster() *schema.Resource {
	r := &schema.Resource{
		Create: resourceUCloudUK8SClusterCreate,
		Read:   resourceUCloudUK8SClusterRead,
		Update: resourceUCloudUK8SClusterUpdate,
		Delete: resourceUCloudUK8SClusterDelete,

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{{
			Version: 0,
			Type:    uk8sClusterStateV0Type(),
			Upgrade: upgradeUK8SClusterStateV0,
		}},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			diffValidateUK8SMaster,
			diffValidateUK8SWorkers,
		),

		Schema: map[string]*schema.Schema{
			"cni_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"VPC", "Calico"}, false),
			},

			"service_cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validateInstancePassword,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"user_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"init_script": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"dynamic",
				}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},

			"k8s_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"enable_external_api_server": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"delete_disks_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"kube_proxy": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"master": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							MinItems: 3,
							MaxItems: 3,
							Required: true,
							ForceNew: true,
						},

						"instance_type": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							ValidateFunc:  validateInstanceType,
							Deprecated:    "Use machine_type, cpu and memory together for new clusters. Existing configurations remain supported.",
							ConflictsWith: []string{"master.0.machine_type", "master.0.cpu", "master.0.memory"},
						},
						"machine_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9]*$`), "must be an uppercase machine type, such as O"),
						},
						"cpu": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(2),
						},
						"memory": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateAll(validation.IntAtLeast(4096), validateMod(1024)),
						},
						"uhost_family": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"o1i", "o2i"}, false),
						},
						"image_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateUImageName,
						},
						"boot_disk_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(40, 500),
						},

						"boot_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local_normal",
								"local_ssd",
								"cloud_normal",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"data_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validateAll(
								validation.IntBetween(20, 1000),
								validateMod(10),
							),
						},

						"data_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local_normal",
								"local_ssd",
								"cloud_normal",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"min_cpu_platform": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "Intel/Auto",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return old == "" && new == "Intel/Auto"
							},
							ValidateFunc: validation.StringInSlice([]string{
								"Intel/Auto",
								"Intel/IvyBridge",
								"Intel/Haswell",
								"Intel/Broadwell",
								"Intel/Skylake",
								"Intel/Cascadelake",
								"Intel/CascadelakeR",
								"Intel/IceLake",
								"Intel/SapphireRapids",
								"Intel/EmeraldRapids",
								"Amd/Auto",
								"Amd/Epyc2",
								"Ampere/Altra",
							}, false),
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"api_server": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"external_api_server": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"pod_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kubeconfig": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"external_kubeconfig": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"image_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateUImageName,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					o, _ := d.GetChange("image_id")
					return o != ""
				},
			},
		},
	}
	r.Schema["worker"] = uk8sClusterWorkerSchema(r.Schema["master"].Elem.(*schema.Resource))
	return r
}

func resourceUCloudUK8SClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating uk8s cluster, %s", err)
	}

	req, err := buildUK8SClusterCreateRequest(client, d)
	if err != nil {
		return err
	}
	var resp sdkuk8s.CreateUK8SClusterV2Response
	err = client.Client.InvokeAction("CreateUK8SClusterV2", req, &resp)
	if err != nil {
		return fmt.Errorf("error on creating uk8s cluster, %s", err)
	}
	if resp.ClusterId == "" {
		return fmt.Errorf("error on creating uk8s cluster, response contains no cluster id")
	}
	d.SetId(resp.ClusterId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{statusPending},
		Target:  []string{statusRUNNING},
		Refresh: func() (interface{}, string, error) {
			instance, err := describeUK8SClusterById(client, d.Id())
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			switch instance.Status {
			case k8sClusterStatusCreateFailed, k8sClusterStatusDeleteFailed, k8sClusterStatusError, k8sClusterStatusAbnormal:
				return instance, "", fmt.Errorf("%s", instance.Status)
			case statusRUNNING:
				return instance, instance.Status, nil
			default:
				return instance, statusPending, nil
			}
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      100 * time.Second,
		MinTimeout: 2 * time.Second,
	}
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for uk8s cluster %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudUK8SClusterRead(d, meta)
}

func resourceUCloudUK8SClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating uk8s cluster, %s", err)
	}
	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := client.NewModifyUK8SClusterNameRequest()
		req.ClusterId = ucloud.String(d.Id())
		req.ClusterName = ucloud.String(d.Get("name").(string))

		if _, err = client.ModifyUK8SClusterName(req); err != nil {
			return fmt.Errorf("error on modifying k8s cluster %q name: %w", d.Id(), err)
		}

		d.SetPartial("name")
	}

	d.Partial(false)
	return resourceUCloudUK8SClusterRead(d, meta)
}

func resourceUCloudUK8SClusterRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading uk8s cluster, %s", err)
	}

	instance, err := describeUK8SClusterById(client, d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading k8s cluster %q, %s", d.Id(), err)
	}

	_ = d.Set("service_cidr", instance.ServiceCIDR)
	_ = d.Set("name", instance.ClusterName)
	_ = d.Set("vpc_id", instance.VPCId)
	_ = d.Set("subnet_id", instance.SubnetId)
	_ = d.Set("status", instance.Status)
	_ = d.Set("create_time", timestampToString(instance.CreateTime))
	_ = d.Set("api_server", instance.ApiServer)
	_ = d.Set("external_api_server", instance.ExternalApiServer)
	_ = d.Set("k8s_version", instance.K8sVersion)
	_ = d.Set("pod_cidr", instance.PodCIDR)
	if instance.CNIMode != "" {
		if err := d.Set("cni_mode", instance.CNIMode); err != nil {
			return fmt.Errorf("error on setting cni_mode for uk8s cluster %q: %w", d.Id(), err)
		}
	}

	configRequest := client.NewGetClusterConfigRequest()
	configRequest.ClusterId = ucloud.String(d.Id())
	config, configErr := client.GetClusterConfig(configRequest)
	if configErr != nil {
		return fmt.Errorf("error on reading kubeconfig for uk8s cluster %q, %s", d.Id(), configErr)
	}
	_ = d.Set("kubeconfig", config.KubeConfig)
	_ = d.Set("external_kubeconfig", config.ExternalKubeConfig)
	return nil
}

func resourceUCloudUK8SClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting uk8s cluster %q: %w", d.Id(), err)
	}

	deleteRequest := client.NewDelUK8SClusterRequest()
	deleteRequest.ClusterId = ucloud.String(d.Id())
	if value, ok := d.GetOkExists("delete_disks_with_instance"); ok {
		deleteRequest.ReleaseUDisk = ucloud.Bool(value.(bool))
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if instance, err := describeUK8SClusterById(client, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading k8s cluster when deleting %q, %s", d.Id(), err))
		} else if instance.Status == "DELETING" {
			return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
		}

		if _, err := client.DelUK8SCluster(deleteRequest); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting k8s cluster %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
	})
}

func resolveUK8SMasterSpec(master map[string]interface{}) (*instanceType, error) {
	if legacy, _ := master["instance_type"].(string); legacy != "" {
		return parseInstanceType(legacy)
	}
	machine, _ := master["machine_type"].(string)
	cpu, _ := master["cpu"].(int)
	memory, _ := master["memory"].(int)
	if machine == "" || cpu < 2 || memory < 4096 || memory%1024 != 0 {
		return nil, fmt.Errorf("master requires instance_type or all of machine_type, cpu >= 2 and memory >= 4096 MB in multiples of 1024")
	}
	return &instanceType{HostType: machine, CPU: cpu, Memory: memory}, nil
}

func diffValidateUK8SMaster(diff *schema.ResourceDiff, meta interface{}) error {
	items := diff.Get("master").([]interface{})
	if len(items) == 0 || items[0] == nil {
		return nil
	}
	master := items[0].(map[string]interface{})
	for _, key := range []string{"instance_type", "machine_type", "cpu", "memory"} {
		if !diff.NewValueKnown("master.0." + key) {
			return nil
		}
	}
	spec, err := resolveUK8SMasterSpec(master)
	if err != nil {
		return err
	}
	legacy := master["instance_type"].(string)
	if legacy != "" {
		// Keep the published validation rules for legacy configurations.
		if err := diffValidateBootDiskTypeWithInstanceTypeOfUK8sCluster(diff, meta); err != nil {
			return err
		}
	}
	machine := strings.ToUpper(spec.HostType)
	family := master["uhost_family"].(string)
	if (legacy == "" || family != "") && diff.NewValueKnown("master.0.boot_disk_type") {
		disk := master["boot_disk_type"].(string)
		if disk == "" {
			disk = "cloud_ssd"
		}
		if strings.HasPrefix(machine, "O") && disk != "cloud_rssd" {
			return fmt.Errorf("master.boot_disk_type must be cloud_rssd for machine_type %s", machine)
		}
	}
	if family != "" {
		if machine != "O" {
			return fmt.Errorf("master.uhost_family %s requires machine_type O", family)
		}
		if diff.NewValueKnown("master.0.min_cpu_platform") && !strings.HasPrefix(master["min_cpu_platform"].(string), "Intel/") {
			return fmt.Errorf("master.uhost_family %s requires an Intel CPU platform", family)
		}
	}
	return nil
}

func buildUK8SClusterCreateRequest(client *sdkuk8s.UK8SClient, d *schema.ResourceData) (request.GenericRequest, error) {
	req := client.NewCreateUK8SClusterV2Request()
	req.ServiceCIDR = ucloud.String(d.Get("service_cidr").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))

	if value, ok := d.GetOk("name"); ok {
		req.ClusterName = ucloud.String(value.(string))
	} else {
		req.ClusterName = ucloud.String(resource.PrefixedUniqueId("tf-uk8s-"))
	}
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))

	if value, ok := d.GetOk("init_script"); ok {
		req.InitScript = ucloud.String(base64.StdEncoding.EncodeToString([]byte(value.(string))))
	}
	if value, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(value.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}
	if value, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(value.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}
	if value, ok := d.GetOk("k8s_version"); ok {
		req.K8sVersion = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("image_id"); ok {
		req.ImageId = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("enable_external_api_server"); ok {
		if value.(bool) {
			req.ExternalApiServer = ucloud.String("Yes")
		} else {
			req.ExternalApiServer = ucloud.String("No")
		}
	}
	if items, ok := d.GetOk("kube_proxy"); ok {
		kubeProxy := items.([]interface{})[0].(map[string]interface{})
		if value, ok := kubeProxy["mode"]; ok {
			req.KubeProxy = &sdkuk8s.CreateUK8SClusterV2ParamKubeProxy{Mode: ucloud.String(value.(string))}
		}
	}

	master := d.Get("master").([]interface{})[0].(map[string]interface{})
	for _, item := range master["availability_zones"].([]interface{}) {
		masterNode := sdkuk8s.CreateUK8SClusterV2ParamMaster{Zone: ucloud.String(item.(string))}
		req.Master = append(req.Master, masterNode)
	}

	parsedInstanceType, err := resolveUK8SMasterSpec(master)
	if err != nil {
		return nil, err
	}
	req.MasterCPU = ucloud.Int(parsedInstanceType.CPU)
	req.MasterMem = ucloud.Int(parsedInstanceType.Memory)
	req.MasterMachineType = ucloud.String(strings.ToUpper(parsedInstanceType.HostType))
	if diskType := master["boot_disk_type"].(string); diskType != "" {
		req.MasterBootDiskType = ucloud.String(upperCvt.unconvert(diskType))
	} else {
		req.MasterBootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
	}
	if diskSize := master["data_disk_size"].(int); diskSize != 0 {
		if diskType := master["data_disk_type"].(string); diskType != "" {
			req.MasterDataDiskType = ucloud.String(upperCvt.unconvert(diskType))
		} else {
			req.MasterDataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
		}
		req.MasterDataDiskSize = ucloud.Int(diskSize)
	}
	if platform := master["min_cpu_platform"].(string); platform != "" {
		req.MasterMinimalCpuPlatform = ucloud.String(platform)
	} else {
		req.MasterMinimalCpuPlatform = ucloud.String("Intel/Auto")
	}

	if value := master["image_id"].(string); value != "" {
		req.MasterImageId = ucloud.String(value)
	}
	if value := master["boot_disk_size"].(int); value != 0 {
		req.MasterBootDiskSize = ucloud.Int(value)
	}
	// The pinned SDK lacks CNIMode and MasterUHostFamily on its create request. Keep its
	// existing parameter encoding and add the field through a generic request.
	payload, err := request.EncodeJSON(req)
	if err != nil {
		return nil, fmt.Errorf("encode uk8s cluster request: %w", err)
	}
	payload["Action"] = "CreateUK8SClusterV2"
	if value, ok := d.GetOk("cni_mode"); ok {
		payload["CNIMode"] = value.(string)
	}
	if value := master["uhost_family"].(string); value != "" {
		payload["MasterUHostFamily"] = value
	}
	workers, err := expandUK8SWorkers(d.Get("worker").([]interface{}))
	if err != nil {
		return nil, err
	}
	if len(workers) > 0 {
		payload["Nodes"] = workers
	}
	generic := client.NewGenericRequest()
	generic.SetRetryable(false)
	if err := generic.SetPayload(payload); err != nil {
		return nil, err
	}
	return generic, nil
}

func diffValidateBootDiskTypeWithInstanceTypeOfUK8sCluster(diff *schema.ResourceDiff, meta interface{}) error {
	master := diff.Get("master").([]interface{})[0].(map[string]interface{})
	instanceType, err := parseInstanceType(master["instance_type"].(string))
	if err != nil {
		return err
	}

	bootDiskType := master["boot_disk_type"].(string)
	if bootDiskType == "" {
		bootDiskType = "cloud_ssd"
	}
	if strings.Contains(instanceType.HostType, "o") && isStringIn(bootDiskType, []string{
		"local_normal",
		"local_ssd",
		"cloud_ssd",
		"cloud_normal",
	}) {
		return fmt.Errorf("the boot_disk_type must be set one of  %v when instance type is belong to outstanding machine , got %q", []string{"cloud_rssd"}, bootDiskType)
	}
	return nil
}
