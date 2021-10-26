package ucloud

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudUK8SCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUK8SClusterCreate,
		Read:   resourceUCloudUK8SClusterRead,
		Update: resourceUCloudUK8SClusterUpdate,
		Delete: resourceUCloudUK8SClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			diffValidateBootDiskTypeWithInstanceTypeOfUK8sCluster,
		),

		Schema: map[string]*schema.Schema{
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
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateInstanceType,
						},

						"boot_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local_normal",
								"local_ssd",
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
								"cloud_normal",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"min_cpu_platform": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Intel/Auto",
								"Intel/IvyBridge",
								"Intel/Haswell",
								"Intel/Broadwell",
								"Intel/Skylake",
								"Intel/Cascadelake",
								"Intel/CascadelakeR",
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
		},
	}
}

func resourceUCloudUK8SClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn

	req := conn.NewCreateUK8SClusterV2Request()
	req.ServiceCIDR = ucloud.String(d.Get("service_cidr").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))

	if v, ok := d.GetOk("name"); ok {
		req.ClusterName = ucloud.String(v.(string))
	} else {
		req.ClusterName = ucloud.String(resource.PrefixedUniqueId("tf-uk8s-"))
	}
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))

	if v, ok := d.GetOk("init_script"); ok {
		req.InitScript = ucloud.String(base64.StdEncoding.EncodeToString([]byte(v.(string))))
	}

	if v, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("k8s_version"); ok {
		req.K8sVersion = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("enable_external_api_server"); ok {
		if v.(bool) {
			req.ExternalApiServer = ucloud.String("Yes")
		} else {
			req.ExternalApiServer = ucloud.String("No")
		}
	}

	if items, ok := d.GetOk("kube_proxy"); ok {
		kubeProxy := items.([]interface{})[0].(map[string]interface{})
		if v, ok := kubeProxy["mode"]; ok {
			req.KubeProxy = &uk8s.CreateUK8SClusterV2ParamKubeProxy{Mode: ucloud.String(v.(string))}
		}
	}

	masterValues := d.Get("master").([]interface{})
	master := masterValues[0].(map[string]interface{})

	for _, item := range master["availability_zones"].([]interface{}) {
		masterNode := uk8s.CreateUK8SClusterV2ParamMaster{}
		masterNode.Zone = ucloud.String(item.(string))
		req.Master = append(req.Master, masterNode)
	}

	// skip error because it has been validated by schema
	mt, _ := parseInstanceType(master["instance_type"].(string))
	req.MasterCPU = ucloud.Int(mt.CPU)
	req.MasterMem = ucloud.Int(mt.Memory)
	req.MasterMachineType = ucloud.String(strings.ToUpper(mt.HostType))

	if len(master["boot_disk_type"].(string)) != 0 {
		req.MasterBootDiskType = ucloud.String(upperCvt.unconvert(master["boot_disk_type"].(string)))
	} else {
		req.MasterBootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
	}

	if master["data_disk_size"].(int) != 0 {
		if len(master["data_disk_type"].(string)) != 0 {
			req.MasterDataDiskType = ucloud.String(upperCvt.unconvert(master["data_disk_type"].(string)))
		} else {
			req.MasterDataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
		}
		req.MasterDataDiskSize = ucloud.Int(master["data_disk_size"].(int))
	}

	if len(master["min_cpu_platform"].(string)) != 0 {
		req.MasterMinmalCpuPlatform = ucloud.String(master["min_cpu_platform"].(string))
	} else {
		req.MasterMinmalCpuPlatform = ucloud.String("Intel/Auto")
	}

	resp, err := conn.CreateUK8SClusterV2(req)
	if err != nil {
		return fmt.Errorf("error on creating uk8s cluster, %s", err)
	}

	d.SetId(resp.ClusterId)

	// after create instance, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusPending},
		Target:  []string{statusRUNNING},
		Refresh: func() (interface{}, string, error) {
			instance, err := client.describeUK8SClusterById(d.Id())
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			switch instance.Status {
			case k8sClusterStatusCreateFailed, k8sClusterStatusDeleteFailed, k8sClusterStatusError, k8sClusterStatusAbnormal:
				return instance, "", fmt.Errorf(instance.Status)
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

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for uk8s cluster %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudUK8SClusterRead(d, meta)
}

func resourceUCloudUK8SClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn
	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewGenericRequest()
		err := req.SetPayload(map[string]interface{}{
			"Action":      "ModifyUK8SClusterName",
			"ClusterId":   d.Id(),
			"ClusterName": d.Get("name").(string),
		})
		if err != nil {
			return fmt.Errorf("error on setting %s request to k8s cluster %q, %s", "ModifyUK8SClusterName", d.Id(), err)
		}
		_, err = conn.GenericInvoke(req)
		if err != nil {
			return fmt.Errorf("error on %s to k8s cluster %q, %s", "ModifyUK8SClusterName", d.Id(), err)
		}

		d.SetPartial("name")
	}

	d.Partial(false)

	return resourceUCloudUK8SClusterRead(d, meta)
}

func resourceUCloudUK8SClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	instance, err := client.describeUK8SClusterById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading k8s cluster %q, %s", d.Id(), err)
	}

	//d.Set("password", d.Get("password"))
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
	return nil
}

func resourceUCloudUK8SClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn

	deleReq := conn.NewDelUK8SClusterRequest()
	deleReq.ClusterId = ucloud.String(d.Id())
	if v, ok := d.GetOkExists("delete_disks_with_instance"); ok {
		deleReq.ReleaseUDisk = ucloud.Bool(v.(bool))
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if inst, err := client.describeUK8SClusterById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading k8s cluster when deleting %q, %s", d.Id(), err))
		} else {
			if inst.Status == "DELETING" {
				return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
			}
		}

		if _, err := conn.DelUK8SCluster(deleReq); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting k8s cluster %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
	})
}

func diffValidateBootDiskTypeWithInstanceTypeOfUK8sCluster(diff *schema.ResourceDiff, meta interface{}) error {
	master := diff.Get("master").([]interface{})[0].(map[string]interface{})
	mt, err := parseInstanceType(master["instance_type"].(string))
	if err != nil {
		return err
	}

	var bootDiskType string
	if len(master["boot_disk_type"].(string)) != 0 {
		bootDiskType = master["boot_disk_type"].(string)
	} else {
		bootDiskType = "cloud_ssd"
	}

	if strings.Contains(mt.HostType, "o") && isStringIn(bootDiskType, []string{
		"local_normal",
		"local_ssd",
		"cloud_ssd",
		"cloud_normal",
	}) {
		return fmt.Errorf("the boot_disk_type must be set one of  %v "+
			"when instance type is belong to outstanding machine , got %q",
			[]string{"cloud_rssd"}, bootDiskType)
	}

	return nil
}
