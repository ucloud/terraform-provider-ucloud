package ucloud

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"time"
)

func resourceUCloudCubePod() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudCubePodCreate,
		Read:   resourceUCloudCubePodRead,
		Update: resourceUCloudCubePodUpdate,
		Delete: resourceUCloudCubePodDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": {
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

			"pod": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateCubePodName,
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"postpay",
				}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"security_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"pod_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudCubePodCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.cubeconn
	req := conn.NewCreateCubePodRequest()

	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.Pod = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("pod").(string))))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
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

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-cube-pod-"))
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	resp, err := conn.CreateCubePod(req)
	if err != nil {
		return fmt.Errorf("error on creating cube pod, %s", err)
	}
	d.SetId(resp.CubeId)

	// after create cube pod, we need to wait it initialized
	stateConf := resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusRunning},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh:    cubePodStateRefreshFunc(client, d.Id(), []string{statusRunning}),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for cube pod %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudCubePodUpdate(d, meta)
}

func resourceUCloudCubePodUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.cubeconn

	d.Partial(true)

	if d.HasChange("pod") && !d.IsNewResource() {
		req := conn.NewRenewCubePodRequest()
		req.Zone = ucloud.String(d.Get("availability_zone").(string))
		req.CubeId = ucloud.String(d.Id())
		req.Pod = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("pod").(string))))

		_, err := conn.RenewCubePod(req)
		if err != nil {
			return fmt.Errorf("error on %s to cube pod %q, %s", "RenewCubePod", d.Id(), err)
		}

		d.SetPartial("pod")
	}

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyCubeExtendInfoRequest()
		req.Zone = ucloud.String(d.Get("availability_zone").(string))
		req.CubeId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		_, err := conn.ModifyCubeExtendInfo(req)
		if err != nil {
			return fmt.Errorf("error on %s to cube pod %q, %s", "ModifyCubeExtendInfo", d.Id(), err)
		}

		d.SetPartial("name")
	}

	if d.HasChange("security_group") {
		conn := client.unetconn
		req := conn.NewGrantFirewallRequest()
		req.FWId = ucloud.String(d.Get("security_group").(string))
		req.ResourceType = ucloud.String(eipResourceTypeCube)
		req.ResourceId = ucloud.String(d.Id())

		_, err := conn.GrantFirewall(req)
		if err != nil {
			return fmt.Errorf("error on %s to cube pod %q, %s", "GrantFirewall", d.Id(), err)
		}

		d.SetPartial("security_group")
	}

	d.Partial(false)

	return resourceUCloudCubePodRead(d, meta)
}

func resourceUCloudCubePodRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	instance, err := client.describeCubePodById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading cube pod %q, %s", d.Id(), err)
	}

	d.Set("availability_zone", instance.CubePodInfo.Metadata.Provider.Zone)
	d.Set("vpc_id", instance.CubePodInfo.Metadata.Provider.VpcId)
	d.Set("subnet_id", instance.CubePodInfo.Metadata.Provider.SubnetId)
	d.Set("tag", instance.CubeExtendInfo.Tag)
	d.Set("name", instance.CubeExtendInfo.Name)
	d.Set("charge_type", upperCamelCvt.convert(instance.CubePodInfo.Metadata.Provider.ChargeType))
	d.Set("expire_time", timestampToString(instance.CubeExtendInfo.Expiration))
	d.Set("status", instance.CubePodInfo.Status.Phase)
	d.Set("pod_ip", instance.CubePodInfo.Status.PodIp)

	if instance.CubePodInfo.Metadata.CreateTimestamp != "" {
		ctStamp, err := stringToTimestamp(instance.CubePodInfo.Metadata.CreateTimestamp)
		if err != nil {
			return fmt.Errorf("error on reading cube pod %q when convert create time, %s", d.Id(), err)
		}
		d.Set("create_time", timestampToString(ctStamp))
	}

	sgSet, err := client.describeFirewallByIdAndType(d.Id(), eipResourceTypeCube)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error on reading security group when reading cube pod %q, %s", d.Id(), err)
	}

	d.Set("security_group", sgSet.FWId)

	return nil
}

func resourceUCloudCubePodDelete(d *schema.ResourceData, meta interface{}) error {
	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		client := meta.(*UCloudClient)
		conn := client.cubeconn

		req := conn.NewDeleteCubePodRequest()
		req.CubeId = ucloud.String(d.Id())

		if _, err := conn.DeleteCubePod(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting cube pod %q, %s", d.Id(), err))
		}

		_, err := client.describeCubePodById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading cube pod when deleting %q, %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified cube pod %q has not been deleted due to unknown error", d.Id()))
	})
}

func cubePodStateRefreshFunc(client *UCloudClient, podId string, target []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pod, err := client.describeCubePodById(podId)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		state := pod.CubePodInfo.Status.Phase
		if !isStringIn(state, target) {
			if state == cubePodStatusCrashLoopBackOff {
				return nil, "", fmt.Errorf("cube pod %q, please make sure the pod image is available", state)
			}
			state = statusPending
		}

		return pod, state, nil
	}
}
