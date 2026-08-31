package uk8s

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
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
					o, _ := d.GetChange("image_id")
					return o != ""
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
					"cloud_normal",
					"cloud_ssd",
					"cloud_rssd",
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating uk8s node, %s", err)
	}

	req := client.NewAddUK8SUHostNodeRequest()
	req.ClusterId = ucloud.String(d.Get("cluster_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Count = ucloud.Int(1)

	if value, ok := d.GetOk("disable_schedule_on_create"); ok {
		req.DisableSchedule = ucloud.Bool(value.(bool))
	}
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))
	if value, ok := d.GetOk("image_id"); ok {
		req.ImageId = ucloud.String(value.(string))
	}
	if value, ok := d.GetOk("user_data"); ok {
		req.UserData = ucloud.String(base64.StdEncoding.EncodeToString([]byte(value.(string))))
	}
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
	if value, ok := d.GetOk("isolation_group"); ok {
		req.IsolationGroup = ucloud.String(value.(string))
	}

	parsedInstanceType, _ := parseInstanceType(d.Get("instance_type").(string))
	req.CPU = ucloud.Int(parsedInstanceType.CPU)
	req.Mem = ucloud.Int(parsedInstanceType.Memory)
	req.MachineType = ucloud.String(strings.ToUpper(parsedInstanceType.HostType))
	if value, ok := d.GetOk("boot_disk_type"); ok {
		req.BootDiskType = ucloud.String(upperCvt.unconvert(value.(string)))
	} else {
		req.BootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
	}
	if value, ok := d.GetOk("data_disk_size"); ok {
		if diskType, ok := d.GetOk("data_disk_type"); ok {
			req.DataDiskType = ucloud.String(upperCvt.unconvert(diskType.(string)))
		} else {
			req.DataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
		}
		req.DataDiskSize = ucloud.Int(value.(int))
	}
	if value, ok := d.GetOk("min_cpu_platform"); ok {
		req.MinmalCpuPlatform = ucloud.String(value.(string))
	} else {
		req.MinmalCpuPlatform = ucloud.String("Intel/Auto")
	}

	resp, err := client.AddUK8SUHostNode(req)
	if err != nil {
		return fmt.Errorf("error on creating uk8s cluster, %s", err)
	}
	if len(resp.NodeIds) == 0 {
		return fmt.Errorf("error on creating uk8s node, response contains no node id")
	}
	d.SetId(resp.NodeIds[0])

	stateConf := &resource.StateChangeConf{
		Pending: []string{statusPending},
		Target:  []string{statusRunning},
		Refresh: func() (interface{}, string, error) {
			node, err := describeUK8SClusterNodeByResourceId(client, d.Get("cluster_id").(string), d.Id())
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}
			switch node.NodeStatus {
			case k8sNodeStatusError, k8sNodeStatusInstallFail, k8sNodeStatusStopped:
				return node, "", fmt.Errorf("%s", node.NodeStatus)
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
	if _, err = stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for uk8s cluster %q complete creating, %s", d.Id(), err)
	}
	return resourceUK8SNodeRead(d, meta)
}

func resourceUK8SNodeUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceUK8SNodeRead(d, meta)
}

func resourceUK8SNodeRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading uk8s node, %s", err)
	}

	clusterID := d.Get("cluster_id").(string)
	node, err := describeUK8SClusterNodeByResourceId(client, clusterID, d.Id())
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
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting uk8s node, %s", err)
	}

	clusterID := d.Get("cluster_id").(string)
	deleteRequest := client.NewDelUK8SClusterNodeV2Request()
	deleteRequest.ClusterId = ucloud.String(clusterID)
	deleteRequest.NodeId = ucloud.String(d.Id())
	if value, ok := d.GetOkExists("delete_disks_with_instance"); ok {
		deleteRequest.ReleaseDataUDisk = ucloud.Bool(value.(bool))
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if node, err := describeUK8SClusterNodeByResourceId(client, clusterID, d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading k8s cluster when deleting %q, %s", d.Id(), err))
		} else if node.NodeStatus == "ToBeDeleted" {
			return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
		}

		if _, err := client.DelUK8SClusterNodeV2(deleteRequest); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting k8s cluster %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
	})
}

func diffValidateBootDiskTypeWithInstanceTypeOfUK8sNode(diff *schema.ResourceDiff, meta interface{}) error {
	instanceType, err := parseInstanceType(diff.Get("instance_type").(string))
	if err != nil {
		return err
	}

	bootDiskType := "cloud_ssd"
	if value, ok := diff.GetOk("boot_disk_type"); ok {
		bootDiskType = value.(string)
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
