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
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudUK8SNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceUK8SNodeCreate,
		Read:   resourceUK8SNodeRead,
		Update: resourceUK8SNodeUpdate,
		Delete: resourceUK8SNodeDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			diffValidateBootDiskTypeWithInstanceTypeOfUK8sNode,
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if o, _ := d.GetChange("image_id"); o != "" {
						return true
					}
					return false
				},
			},

			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validateInstancePassword,
			},

			"instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateInstanceType,
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
				ValidateFunc: validateAll(
					validation.IntBetween(0, 2000),
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
				}, false),
			},

			"isolation_group": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

			"delete_disks_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"disable_schedule_on_create": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
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

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"internet_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUK8SNodeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn

	req := conn.NewAddUK8SUHostNodeRequest()

	req.ClusterId = ucloud.String(d.Get("cluster_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Count = ucloud.Int(1)

	if v, ok := d.GetOk("disable_schedule_on_create"); ok {
		req.DisableSchedule = ucloud.Bool(v.(bool))
	}

	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))

	if v, ok := d.GetOk("image_id"); ok {
		req.ImageId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("user_data"); ok {
		req.UserData = ucloud.String(base64.StdEncoding.EncodeToString([]byte(v.(string))))
	}

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

	if v, ok := d.GetOk("isolation_group"); ok {
		req.IsolationGroup = ucloud.String(v.(string))
	}

	// skip error because it has been validated by schema
	mt, _ := parseInstanceType(d.Get("instance_type").(string))
	req.CPU = ucloud.Int(mt.CPU)
	req.Mem = ucloud.Int(mt.Memory)
	req.MachineType = ucloud.String(strings.ToUpper(mt.HostType))

	if v, ok := d.GetOk("boot_disk_type"); ok {
		req.BootDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
	} else {
		req.BootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
	}

	if v, ok := d.GetOk("data_disk_size"); ok {
		if val, ok := d.GetOk("data_disk_type"); ok {
			req.DataDiskType = ucloud.String(upperCvt.unconvert(val.(string)))
		} else {
			req.DataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
		}
		req.DataDiskSize = ucloud.Int(v.(int))
	}

	if v, ok := d.GetOk("min_cpu_platform"); ok {
		req.MinmalCpuPlatform = ucloud.String(v.(string))
	} else {
		req.MinmalCpuPlatform = ucloud.String("Intel/Auto")
	}

	if v, ok := d.GetOk("max_pods"); ok {
		req.MaxPods = ucloud.Int(v.(int))
	}

	if v, ok := d.GetOk("labels"); ok {
		var labelList []string
		labels := v.([]interface{})
		for _, label := range labels {
			v := label.(map[string]interface{})
			labelList = append(labelList, fmt.Sprintf("%s=%s", v["key"], v["value"]))
		}
		req.Labels = ucloud.String(strings.Join(labelList, ","))
	}

	resp, err := conn.AddUK8SUHostNode(req)
	if err != nil {
		return fmt.Errorf("error on creating uk8s cluster, %s", err)
	}

	d.SetId(resp.NodeIds[0])

	// after create instance, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusPending},
		Target:  []string{statusRunning},
		Refresh: func() (interface{}, string, error) {
			node, err := client.describeUK8SClusterNodeByResourceId(d.Get("cluster_id").(string), d.Id())
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			switch node.NodeStatus {
			case k8sNodeStatusError, k8sNodeStatusInstallFail, k8sNodeStatusStopped:
				return node, "", fmt.Errorf(node.NodeStatus)
			case k8sNodeStatusReady:
				return node, statusRunning, nil
			default:
				return node, statusPending, nil
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
	return resourceUK8SNodeRead(d, meta)
}

func resourceUK8SNodeUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceUK8SNodeRead(d, meta)
}

func resourceUK8SNodeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	clusterId := d.Get("cluster_id").(string)
	node, err := client.describeUK8SClusterNodeByResourceId(clusterId, d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading k8s cluster %q, %s", d.Id(), err)
	}

	var ipSet []map[string]interface{}
	for _, item := range node.IPSet {
		ipSet = append(ipSet, map[string]interface{}{
			"ip":            item.IP,
			"internet_type": item.Type,
		})
	}
	_ = d.Set("ip_set", ipSet)
	_ = d.Set("status", node.NodeStatus)
	_ = d.Set("create_time", timestampToString(node.CreateTime))
	_ = d.Set("expire_time", timestampToString(node.ExpireTime))
	return nil
}

func resourceUK8SNodeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn

	clusterId := d.Get("cluster_id").(string)
	deleReq := conn.NewDelUK8SClusterNodeV2Request()
	deleReq.ClusterId = ucloud.String(clusterId)
	deleReq.NodeId = ucloud.String(d.Id())
	if v, ok := d.GetOkExists("delete_disks_with_instance"); ok {
		deleReq.ReleaseDataUDisk = ucloud.Bool(v.(bool))
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if inst, err := client.describeUK8SClusterNodeByResourceId(clusterId, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading k8s cluster when deleting %q, %s", d.Id(), err))
		} else {
			if inst.NodeStatus == "ToBeDeleted" {
				return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
			}
		}
		if _, err := conn.DelUK8SClusterNodeV2(deleReq); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting k8s cluster %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
	})
}

func diffValidateBootDiskTypeWithInstanceTypeOfUK8sNode(diff *schema.ResourceDiff, meta interface{}) error {
	mt, err := parseInstanceType(diff.Get("instance_type").(string))
	if err != nil {
		return err
	}

	var bootDiskType string
	if v, ok := diff.GetOk("boot_disk_type"); ok {
		bootDiskType = v.(string)
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
