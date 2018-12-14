package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudInstanceCreate,
		Read:   resourceUCloudInstanceRead,
		Update: resourceUCloudInstanceUpdate,
		Delete: resourceUCloudInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"image_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"root_password": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validateInstancePassword,
			},

			"instance_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateInstanceType,
			},

			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      resource.PrefixedUniqueId("tf-instance-"),
				ValidateFunc: validateName,
			},

			"charge_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "month",
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"dynamic",
				}, false),
			},

			"duration": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1,
				ValidateFunc: validateDuration,
			},

			"boot_disk_size": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(20, 100),
			},

			"boot_disk_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "local_normal",
				ValidateFunc: validation.StringInSlice([]string{"local_normal", "local_ssd", "cloud_normal", "cloud_ssd"}, false),
			},

			"data_disk_size": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},

			"data_disk_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "local_normal",
				ValidateFunc: validation.StringInSlice([]string{"local_normal", "local_ssd"}, false),
			},

			"remark": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tag": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateTag,
			},

			"security_group": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"disk_set": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"size": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"is_boot": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"ip_set": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"internet_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"create_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"expire_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_renew": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceUCloudInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uhostconn

	imageId := d.Get("image_id").(string)
	bootDiskType := d.Get("boot_disk_type").(string)

	req := conn.NewCreateUHostInstanceRequest()
	req.LoginMode = ucloud.String("Password")
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.ImageId = ucloud.String(imageId)
	req.Password = ucloud.String(d.Get("root_password").(string))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Quantity = ucloud.Int(d.Get("duration").(int))
	req.Name = ucloud.String(d.Get("name").(string))

	// skip error because it has been validated by schema
	t, _ := parseInstanceType(d.Get("instance_type").(string))
	req.CPU = ucloud.Int(t.CPU)
	req.Memory = ucloud.Int(t.Memory)

	bootDisk := uhost.UHostDisk{}
	imageResp, err := client.DescribeImageById(imageId)
	if err != nil {
		return fmt.Errorf("error on reading image %s when creating instance, %s", imageId, err)
	}

	if v, ok := d.GetOk("boot_disk_size"); ok && (bootDiskType == "cloud_normal" || bootDiskType == "cloud_ssd") {
		if v.(int) < imageResp.ImageSize {
			return fmt.Errorf("expected boot_disk_size to be at least %d", imageResp.ImageSize)
		}
		bootDisk.IsBoot = ucloud.String("True")
		bootDisk.Size = ucloud.Int(v.(int))
		bootDisk.Type = ucloud.String(upperCvt.unconvert(bootDiskType))
		req.Disks = append(req.Disks, bootDisk)
	} else {
		bootDisk.IsBoot = ucloud.String("True")
		bootDisk.Size = ucloud.Int(imageResp.ImageSize)
		bootDisk.Type = ucloud.String(upperCvt.unconvert(bootDiskType))
		req.Disks = append(req.Disks, bootDisk)
	}

	if v, ok := d.GetOk("data_disk_size"); ok {
		dataDisk := uhost.UHostDisk{}
		dataDisk.IsBoot = ucloud.String("False")
		dataDisk.Type = ucloud.String(upperCvt.unconvert(d.Get("data_disk_type").(string)))
		dataDisk.Size = ucloud.Int(v.(int))

		req.Disks = append(req.Disks, dataDisk)
	}

	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(v.(string))
	}

	if val, ok := d.GetOk("security_group"); ok {
		resp, err := client.describeFirewallById(val.(string))
		if err != nil {
			return fmt.Errorf("error on reading security group %s when creating instance, %s", val.(string), err)
		}

		req.SecurityGroupId = ucloud.String(resp.GroupId)
	}

	resp, err := conn.CreateUHostInstance(req)
	if err != nil {
		return fmt.Errorf("error on creating instance, %s", err)
	}

	if len(resp.UHostIds) != 1 {
		return fmt.Errorf("error on creating instance, expected exactly one instance, got %v", len(resp.UHostIds))
	}

	d.SetId(resp.UHostIds[0])

	// after create instance, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusRunning},
		Refresh:    instanceStateRefreshFunc(client, d.Id(), statusRunning),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for instance %s complete creating, %s", d.Id(), err)
	}

	return resourceUCloudInstanceUpdate(d, meta)
}

func resourceUCloudInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uhostconn
	d.Partial(true)

	if d.HasChange("security_group") && !d.IsNewResource() {
		conn := client.unetconn
		req := conn.NewGrantFirewallRequest()
		req.FWId = ucloud.String(d.Get("security_group").(string))
		req.ResourceType = ucloud.String("UHost")
		req.ResourceId = ucloud.String(d.Id())

		_, err := conn.GrantFirewall(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %s, %s", "GrantFirewall", d.Id(), err)
		}

		d.SetPartial("security_group")
	}

	if d.HasChange("remark") {
		req := conn.NewModifyUHostInstanceRemarkRequest()
		req.UHostId = ucloud.String(d.Id())
		req.Remark = ucloud.String(d.Get("remark").(string))

		_, err := conn.ModifyUHostInstanceRemark(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %s, %s", "ModifyUHostInstanceRemark", d.Id(), err)
		}

		d.SetPartial("remark")
	}

	if d.HasChange("tag") && !d.IsNewResource() {
		req := conn.NewModifyUHostInstanceTagRequest()
		req.UHostId = ucloud.String(d.Id())
		req.Tag = ucloud.String(d.Get("tag").(string))

		_, err := conn.ModifyUHostInstanceTag(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %s, %s", "ModifyUHostInstanceTag", d.Id(), err)
		}

		d.SetPartial("tag")
	}

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyUHostInstanceNameRequest()
		req.UHostId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		_, err := conn.ModifyUHostInstanceName(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %s, %s", "ModifyUHostInstanceName", d.Id(), err)
		}

		d.SetPartial("name")
	}

	resizeNeedUpdate := false
	resizeReq := conn.NewResizeUHostInstanceRequest()
	resizeReq.UHostId = ucloud.String(d.Id())
	if d.HasChange("instance_type") && !d.IsNewResource() {
		oldType, newType := d.GetChange("instance_type")

		oldInstanceType, _ := parseInstanceType(oldType.(string))
		newInstanceType, _ := parseInstanceType(newType.(string))

		if oldInstanceType.CPU != newInstanceType.CPU {
			resizeReq.CPU = ucloud.Int(newInstanceType.CPU)
		}

		if oldInstanceType.Memory != newInstanceType.Memory {
			resizeReq.Memory = ucloud.Int(newInstanceType.Memory)
		}

		resizeNeedUpdate = true
	}

	if d.HasChange("data_disk_size") && !d.IsNewResource() {
		oldSize, newSize := d.GetChange("data_disk_size")
		if oldSize.(int) > newSize.(int) {
			return fmt.Errorf("reduce data disk size is not supported, new value %d should be larger than the old value %d", newSize.(int), oldSize.(int))
		}
		resizeReq.DiskSpace = ucloud.Int(newSize.(int))
		resizeNeedUpdate = true
	}

	if d.HasChange("boot_disk_size") {
		imageResp, err := client.DescribeImageById(d.Get("image_id").(string))
		if err != nil {
			return fmt.Errorf("error on %s when updating instance %s, %s", "DescribeImage", d.Id(), err)
		}

		if d.Get("boot_disk_size").(int) < imageResp.ImageSize {
			return fmt.Errorf("expected boot_disk_size to be at least %d", imageResp.ImageSize)
		}

		oldSize, newSize := d.GetChange("boot_disk_size")
		if oldSize.(int) > newSize.(int) {
			return fmt.Errorf("reduce boot disk size is not supported, new value %d by user set should be larger than the old value %d allocated by the system", newSize.(int), oldSize.(int))
		}

		resizeReq.BootDiskSpace = ucloud.Int(newSize.(int))
		resizeNeedUpdate = true
	}

	passwordNeedUpdate := false
	if d.HasChange("root_password") && !d.IsNewResource() {
		instance, err := client.describeInstanceById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading instance when updating %s, %s", d.Id(), err)
		}

		if instance.BootDiskState == "Normal" {
			passwordNeedUpdate = true
		} else {
			return fmt.Errorf("reset password not successful, please try again 20 minutes later after the instance starts up")
		}
	}

	if passwordNeedUpdate || resizeNeedUpdate {
		// instance update these attributes need to wait it stopped
		stopReq := conn.NewStopUHostInstanceRequest()
		stopReq.UHostId = ucloud.String(d.Id())

		instance, err := client.describeInstanceById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading instance when updating %s, %s", d.Id(), err)
		}

		if strings.ToLower(instance.State) != statusStopped {
			_, err := conn.StopUHostInstance(stopReq)
			if err != nil {
				return fmt.Errorf("error on stopping instance when updating %s, %s", d.Id(), err)
			}

			// after stop instance, we need to wait it stopped
			stateConf := &resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{statusStopped},
				Refresh:    instanceStateRefreshFunc(client, d.Id(), statusStopped),
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				Delay:      5 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error on waiting for stopping instance when updating %s, %s", d.Id(), err)
			}
		}

		if passwordNeedUpdate {
			reqPassword := conn.NewResetUHostInstancePasswordRequest()
			reqPassword.UHostId = ucloud.String(d.Id())
			reqPassword.Password = ucloud.String(d.Get("root_password").(string))

			_, err := conn.ResetUHostInstancePassword(reqPassword)
			if err != nil {
				return fmt.Errorf("error on %s to instance %s, %s", "ResetUHostInstancePassword", d.Id(), err)
			}

			d.SetPartial("root_password")
		}

		if resizeNeedUpdate {
			_, err := conn.ResizeUHostInstance(resizeReq)
			if err != nil {
				return fmt.Errorf("error on %s to instance %s, %s", "ResizeUHostInstance", d.Id(), err)
			}
		}

		d.SetPartial("instance_type")
		d.SetPartial("boot_disk_size")
		d.SetPartial("data_disk_size")

		// instance stopped means instance update complete
		stateConf := &resource.StateChangeConf{
			Pending:    []string{statusPending},
			Target:     []string{statusStopped},
			Refresh:    instanceStateRefreshFunc(client, d.Id(), statusStopped),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		if _, err = stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for %s complete to instance %s, %s", "ResizeUHostInstance", d.Id(), err)
		}

		if strings.ToLower(instance.State) == statusRunning {
			// after instance update, we need to wait it started
			startReq := conn.NewStartUHostInstanceRequest()
			startReq.UHostId = ucloud.String(d.Id())

			if _, err := conn.StartUHostInstance(startReq); err != nil {
				return fmt.Errorf("error on starting instance when updating %s, %s", d.Id(), err)
			}

			stateConf = &resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{statusRunning},
				Refresh:    instanceStateRefreshFunc(client, d.Id(), statusRunning),
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				Delay:      5 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error on waiting for starting instance when updating %s, %s", d.Id(), err)
			}
		}
	}

	d.Partial(false)

	return resourceUCloudInstanceRead(d, meta)
}

func resourceUCloudInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	instance, err := client.describeInstanceById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading instance %s, %s", d.Id(), err)
	}

	d.Set("name", instance.Name)
	d.Set("charge_type", upperCamelCvt.convert(instance.ChargeType))
	d.Set("availability_zone", instance.Zone)
	d.Set("instance_type", d.Get("instance_type").(string))
	d.Set("root_password", d.Get("root_password").(string))
	d.Set("security_group", d.Get("security_group").(string))
	d.Set("tag", instance.Tag)
	d.Set("cpu", instance.CPU)
	d.Set("memory", instance.Memory)
	d.Set("status", strings.Replace(instance.State, " ", "", -1))
	d.Set("create_time", timestampToString(instance.CreateTime))
	d.Set("expire_time", timestampToString(instance.ExpireTime))
	d.Set("auto_renew", boolCamelCvt.unconvert(instance.AutoRenew))
	d.Set("remark", instance.Remark)

	ipSet := []map[string]interface{}{}
	for _, item := range instance.IPSet {
		ipSet = append(ipSet, map[string]interface{}{
			"ip":            item.IP,
			"internet_type": item.Type,
		})

		if item.Type == "Private" {
			d.Set("vpc_id", item.VPCId)
			d.Set("subnet_id", item.SubnetId)
		}
	}

	if err := d.Set("ip_set", ipSet); err != nil {
		return err
	}

	diskSet := []map[string]interface{}{}
	for _, item := range instance.DiskSet {
		diskSet = append(diskSet, map[string]interface{}{
			"type":    upperCvt.convert(item.DiskType),
			"size":    item.Size,
			"id":      item.DiskId,
			"is_boot": boolValueCvt.unconvert(item.IsBoot),
		})

		isBoot := boolValueCvt.unconvert(item.IsBoot)
		if isBoot {
			d.Set("boot_disk_size", item.Size)
		}

		if !isBoot && checkStringIn(upperCvt.convert(item.DiskType), []string{"local_normal", "local_ssd"}) == nil {
			d.Set("data_disk_size", item.Size)
		}
	}

	if err := d.Set("disk_set", diskSet); err != nil {
		return err
	}

	return nil
}

func resourceUCloudInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uhostconn

	stopReq := conn.NewStopUHostInstanceRequest()
	stopReq.UHostId = ucloud.String(d.Id())

	deleReq := conn.NewTerminateUHostInstanceRequest()
	deleReq.UHostId = ucloud.String(d.Id())

	return resource.Retry(15*time.Minute, func() *resource.RetryError {
		instance, err := client.describeInstanceById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if strings.ToLower(instance.State) != statusStopped {
			if _, err := conn.StopUHostInstance(stopReq); err != nil {
				return resource.RetryableError(fmt.Errorf("error on stopping instance when deleting %s, %s", d.Id(), err))
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{statusStopped},
				Refresh:    instanceStateRefreshFunc(client, d.Id(), statusStopped),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				Delay:      5 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return resource.RetryableError(fmt.Errorf("error on waiting for stopping instance when deleting %s, %s", d.Id(), err))
			}
		}

		if _, err := conn.TerminateUHostInstance(deleReq); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting instance %s, %s", d.Id(), err))
		}

		if _, err := client.describeInstanceById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error on reading instance when deleting %s, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified instance %s has not been deleted due to unknown error", d.Id()))
	})
}

func instanceStateRefreshFunc(client *UCloudClient, instanceId, target string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := client.describeInstanceById(instanceId)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		state := strings.ToLower(instance.State)
		if state != target {
			state = statusPending
		}

		return instance, state, nil
	}
}
