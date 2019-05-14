package ucloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"

	"github.com/hashicorp/terraform/helper/customdiff"
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			// if no default security group under this account, check it to trigger creating event
			customdiff.ValidateChange("security_group", diffValidateDefaultSecurityGroup),
			customdiff.ValidateChange("data_disk_size", diffValidateInstanceDataDiskSize),
			customdiff.ValidateChange("boot_disk_size", diffValidateInstanceBootDiskSize),
			diffValidateBootDiskTypeWithDataDiskType,
			diffValidateInstanceTypeWithZone,
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if o, _ := d.GetChange("image_id"); o != "" {
						return true
					}
					return false
				},
			},

			"root_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ValidateFunc: validateInstancePassword,
			},

			"instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateInstanceType,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "month",
				ForceNew: true,
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

			"boot_disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validateAll(
					validation.IntBetween(20, 100),
					validateMod(10),
				),
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

			"remark": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultTag,
				ValidateFunc: validateTag,
				StateFunc:    stateFuncTag,
			},

			"security_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disk_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"is_boot": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
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

			"private_ip": {
				Type:     schema.TypeString,
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

			"auto_renew": {
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
	var bootDiskType string
	if v, ok := d.GetOk("boot_disk_type"); ok {
		bootDiskType = v.(string)
	} else {
		bootDiskType = "local_normal"
	}
	// skip error because it has been validated by schema
	t, _ := parseInstanceType(d.Get("instance_type").(string))

	req := conn.NewCreateUHostInstanceRequest()
	req.LoginMode = ucloud.String("Password")
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.ImageId = ucloud.String(imageId)
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.CPU = ucloud.Int(t.CPU)
	req.Memory = ucloud.Int(t.Memory)
	password := fmt.Sprintf("%s%s", acctest.RandStringFromCharSet(5, defaultPasswordStr), acctest.RandStringFromCharSet(5, defaultPasswordNum))
	if v, ok := d.GetOk("root_password"); ok {
		req.Password = ucloud.String(v.(string))
	} else {
		req.Password = ucloud.String(password)
	}

	if t.HostType == "o" {
		req.MachineType = ucloud.String("O")
		req.MinimalCpuPlatform = ucloud.String("Intel/Cascadelake")
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-instance-"))
	}

	if v, ok := d.GetOk("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	imageResp, err := client.DescribeImageById(imageId)
	if err != nil {
		return fmt.Errorf("error on reading image %q when creating instance, %s", imageId, err)
	}
	bootDisk := uhost.UHostDisk{}
	bootDisk.IsBoot = ucloud.String("true")
	bootDisk.Type = ucloud.String(upperCvt.unconvert(bootDiskType))
	bootDisk.Size = ucloud.Int(imageResp.ImageSize)

	if v, ok := d.GetOk("boot_disk_size"); ok {
		if v.(int) < imageResp.ImageSize {
			return fmt.Errorf("expected boot_disk_size to be at least %d", imageResp.ImageSize)
		}

		if bootDiskType == "cloud_normal" || bootDiskType == "cloud_ssd" {
			bootDisk.Size = ucloud.Int(v.(int))
		}
	}

	req.Disks = append(req.Disks, bootDisk)

	if v, ok := d.GetOk("data_disk_size"); ok {
		dataDisk := uhost.UHostDisk{}
		dataDisk.IsBoot = ucloud.String("false")
		if val, ok := d.GetOk("data_disk_type"); ok {
			dataDisk.Type = ucloud.String(upperCvt.unconvert(val.(string)))
		} else {
			dataDisk.Type = ucloud.String("LOCAL_NORMAL")
		}
		dataDisk.Size = ucloud.Int(v.(int))

		req.Disks = append(req.Disks, dataDisk)
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
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
			return fmt.Errorf("error on reading security group %q when creating instance, %s", val.(string), err)
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

	if _, ok := d.GetOk("root_password"); !ok {
		d.Set("root_password", password)
	}
	// after create instance, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusRunning},
		Refresh:    instanceStateRefreshFunc(client, d.Id(), statusRunning),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for instance %q complete creating, %s", d.Id(), err)
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
			return fmt.Errorf("error on %s to instance %q, %s", "GrantFirewall", d.Id(), err)
		}

		d.SetPartial("security_group")
	}

	if d.HasChange("remark") {
		req := conn.NewModifyUHostInstanceRemarkRequest()
		req.UHostId = ucloud.String(d.Id())
		req.Remark = ucloud.String(d.Get("remark").(string))

		_, err := conn.ModifyUHostInstanceRemark(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %q, %s", "ModifyUHostInstanceRemark", d.Id(), err)
		}

		d.SetPartial("remark")
	}

	if d.HasChange("tag") && !d.IsNewResource() {
		req := conn.NewModifyUHostInstanceTagRequest()
		req.UHostId = ucloud.String(d.Id())

		// if tag is empty string, use default tag
		if v, ok := d.GetOk("tag"); ok {
			req.Tag = ucloud.String(v.(string))
		} else {
			req.Tag = ucloud.String(defaultTag)
		}

		_, err := conn.ModifyUHostInstanceTag(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %q, %s", "ModifyUHostInstanceTag", d.Id(), err)
		}

		d.SetPartial("tag")
	}

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyUHostInstanceNameRequest()
		req.UHostId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		_, err := conn.ModifyUHostInstanceName(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %q, %s", "ModifyUHostInstanceName", d.Id(), err)
		}

		d.SetPartial("name")
	}

	resizeNeedUpdate := false
	dataDiskNeedUpdate := false
	bootDiskNeedUpdate := false
	zone := d.Get("availability_zone").(string)
	resizeReq := conn.NewResizeUHostInstanceRequest()
	resizeReq.UHostId = ucloud.String(d.Id())
	dataDiskReq := conn.NewResizeAttachedDiskRequest()
	dataDiskReq.UHostId = ucloud.String(d.Id())
	dataDiskReq.Zone = ucloud.String(zone)
	bootDiskReq := conn.NewResizeAttachedDiskRequest()
	bootDiskReq.UHostId = ucloud.String(d.Id())
	bootDiskReq.Zone = ucloud.String(zone)
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
		dataDiskReq.DiskSpace = ucloud.Int(d.Get("data_disk_size").(int))
		dataDiskReq.UHostId = ucloud.String(d.Id())
		dataDiskNeedUpdate = true
	}

	if d.HasChange("boot_disk_size") {
		bootDiskType := d.Get("boot_disk_type").(string)
		imageResp, err := client.DescribeImageById(d.Get("image_id").(string))
		if err != nil {
			return fmt.Errorf("error on %s when updating instance %q, %s", "DescribeImage", d.Id(), err)
		}

		if d.Get("boot_disk_size").(int) < imageResp.ImageSize {
			return fmt.Errorf("expected boot_disk_size to be at least %d", imageResp.ImageSize)
		}

		// the initialization of cloud boot disk is done at creation instance
		if (bootDiskType == "cloud_normal" || bootDiskType == "cloud_ssd") && d.IsNewResource() {
			bootDiskNeedUpdate = false
		} else {
			bootDiskReq.DiskSpace = ucloud.Int(d.Get("boot_disk_size").(int))
			bootDiskReq.UHostId = ucloud.String(d.Id())
			bootDiskNeedUpdate = true
		}
	}

	passwordNeedUpdate := false
	if d.HasChange("root_password") && !d.IsNewResource() {
		instance, err := client.describeInstanceById(d.Id())

		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading instance when updating %q, %s", d.Id(), err)
		}

		if instance.BootDiskState == "Normal" {
			passwordNeedUpdate = true
		} else {
			return fmt.Errorf("reset password not successful, please try again 20 minutes later after the instance starts up")
		}
	}

	if passwordNeedUpdate || resizeNeedUpdate || dataDiskNeedUpdate || bootDiskNeedUpdate {
		// instance update these attributes need to wait it stopped
		stopReq := conn.NewStopUHostInstanceRequest()
		stopReq.UHostId = ucloud.String(d.Id())

		instance, err := client.describeInstanceById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading instance when updating %q, %s", d.Id(), err)
		}

		if strings.ToLower(instance.State) != statusStopped {
			_, err := conn.StopUHostInstance(stopReq)
			if err != nil {
				return fmt.Errorf("error on stopping instance when updating %q, %s", d.Id(), err)
			}

			// after stop instance, we need to wait it stopped
			stateConf := &resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{statusStopped},
				Refresh:    instanceStateRefreshFunc(client, d.Id(), statusStopped),
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				Delay:      3 * time.Second,
				MinTimeout: 2 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error on waiting for stopping instance when updating %q, %s", d.Id(), err)
			}
		}

		if passwordNeedUpdate {
			reqPassword := conn.NewResetUHostInstancePasswordRequest()
			reqPassword.UHostId = ucloud.String(d.Id())
			reqPassword.Password = ucloud.String(d.Get("root_password").(string))

			_, err := conn.ResetUHostInstancePassword(reqPassword)
			if err != nil {
				return fmt.Errorf("error on %s to instance %q, %s", "ResetUHostInstancePassword", d.Id(), err)
			}

			d.SetPartial("root_password")
		}

		if resizeNeedUpdate {
			_, err := conn.ResizeUHostInstance(resizeReq)
			if err != nil {
				return fmt.Errorf("error on %s to instance %q, %s", "ResizeUHostInstance", d.Id(), err)
			}

			d.SetPartial("instance_type")
		}

		if dataDiskNeedUpdate {
			instance, err := client.describeInstanceById(d.Id())
			for _, item := range instance.DiskSet {
				diskType := upperCvt.convert(item.DiskType)
				isBoot := boolValueCvt.unconvert(item.IsBoot)

				if !isBoot && checkStringIn(diskType, []string{"local_normal", "local_ssd"}) == nil {
					dataDiskReq.DiskId = ucloud.String(item.DiskId)
					break
				}
			}

			_, err = conn.ResizeAttachedDisk(dataDiskReq)
			if err != nil {
				return fmt.Errorf("error on %s to instance %q, %s", "ResizeAttachedDisk", d.Id(), err)
			}

			d.SetPartial("data_disk_size")
		}

		if bootDiskNeedUpdate {
			instance, err := client.describeInstanceById(d.Id())
			for _, item := range instance.DiskSet {
				isBoot := boolValueCvt.unconvert(item.IsBoot)

				if isBoot {
					bootDiskReq.DiskId = ucloud.String(item.DiskId)
					break
				}
			}

			_, err = conn.ResizeAttachedDisk(bootDiskReq)
			if err != nil {
				return fmt.Errorf("error on %s to instance %q, %s", "ResizeAttachedDisk", d.Id(), err)
			}

			d.SetPartial("boot_disk_size")
		}

		// instance stopped means instance update complete
		stateConf := &resource.StateChangeConf{
			Pending:    []string{statusPending},
			Target:     []string{statusStopped},
			Refresh:    instanceStateRefreshFunc(client, d.Id(), statusStopped),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 2 * time.Second,
		}

		if _, err = stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for %s complete to instance %q, %s", "ResizeUHostInstance", d.Id(), err)
		}

		if strings.ToLower(instance.State) == statusRunning {
			// after instance update, we need to wait it started
			startReq := conn.NewStartUHostInstanceRequest()
			startReq.UHostId = ucloud.String(d.Id())

			if _, err := conn.StartUHostInstance(startReq); err != nil {
				return fmt.Errorf("error on starting instance when updating %q, %s", d.Id(), err)
			}

			stateConf = &resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{statusRunning},
				Refresh:    instanceStateRefreshFunc(client, d.Id(), statusRunning),
				Timeout:    d.Timeout(schema.TimeoutUpdate),
				Delay:      3 * time.Second,
				MinTimeout: 2 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error on waiting for starting instance when updating %q, %s", d.Id(), err)
			}
		}
	}

	d.Partial(false)

	return resourceUCloudInstanceRead(d, meta)
}

func resourceUCloudInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.unetconn

	instance, err := client.describeInstanceById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading instance %q, %s", d.Id(), err)
	}

	memory := instance.Memory
	cpu := instance.CPU
	d.Set("root_password", d.Get("root_password"))
	d.Set("name", instance.Name)
	d.Set("charge_type", upperCamelCvt.convert(instance.ChargeType))
	d.Set("availability_zone", instance.Zone)
	d.Set("tag", instance.Tag)
	d.Set("cpu", cpu)
	d.Set("memory", memory/1024)
	d.Set("status", instance.State)
	d.Set("create_time", timestampToString(instance.CreateTime))
	d.Set("expire_time", timestampToString(instance.ExpireTime))
	d.Set("auto_renew", boolCamelCvt.unconvert(instance.AutoRenew))
	d.Set("remark", instance.Remark)
	d.Set("instance_type", instanceTypeSetFunc(upperCvt.convert(instance.MachineType), cpu, memory/1024))

	basicImageId := instance.BasicImageId
	if basicImageId != "" {
		d.Set("image_id", basicImageId)
	}

	ipSet := []map[string]interface{}{}
	for _, item := range instance.IPSet {
		ipSet = append(ipSet, map[string]interface{}{
			"ip":            item.IP,
			"internet_type": item.Type,
		})

		if item.Type == "Private" {
			d.Set("vpc_id", item.VPCId)
			d.Set("subnet_id", item.SubnetId)
			d.Set("private_ip", item.IP)
		}
	}

	if err := d.Set("ip_set", ipSet); err != nil {
		return err
	}

	diskSet := []map[string]interface{}{}
	for _, item := range instance.DiskSet {
		diskType := upperCvt.convert(item.DiskType)
		isBoot := boolValueCvt.unconvert(item.IsBoot)
		diskSet = append(diskSet, map[string]interface{}{
			"type":    diskType,
			"size":    item.Size,
			"id":      item.DiskId,
			"is_boot": isBoot,
		})

		if isBoot {
			d.Set("boot_disk_size", item.Size)
			d.Set("boot_disk_type", diskType)
		}

		if !isBoot && checkStringIn(diskType, []string{"local_normal", "local_ssd"}) == nil {
			d.Set("data_disk_size", item.Size)
			d.Set("data_disk_type", diskType)
		}
	}

	if err := d.Set("disk_set", diskSet); err != nil {
		return err
	}

	req := conn.NewDescribeFirewallRequest()
	req.ResourceId = ucloud.String(d.Id())
	req.ResourceType = ucloud.String(eipResourceTypeUHost)
	resp, err := conn.DescribeFirewall(req)

	if err != nil {
		return fmt.Errorf("error on reading security group when reading instance %q, %s", d.Id(), err)
	}
	d.Set("security_group", resp.DataSet[0].FWId)

	return nil
}

func resourceUCloudInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uhostconn

	stopReq := conn.NewPoweroffUHostInstanceRequest()
	stopReq.UHostId = ucloud.String(d.Id())

	deleReq := conn.NewTerminateUHostInstanceRequest()
	deleReq.UHostId = ucloud.String(d.Id())
	deleReq.ReleaseUDisk = ucloud.Bool(true)
	deleReq.ReleaseEIP = ucloud.Bool(false)

	return resource.Retry(15*time.Minute, func() *resource.RetryError {
		instance, err := client.describeInstanceById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if strings.ToLower(instance.State) != statusStopped {
			if _, err := conn.PoweroffUHostInstance(stopReq); err != nil {
				return resource.RetryableError(fmt.Errorf("error on stopping instance when deleting %q, %s", d.Id(), err))
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{statusStopped},
				Refresh:    instanceStateRefreshFunc(client, d.Id(), statusStopped),
				Timeout:    d.Timeout(schema.TimeoutDelete),
				Delay:      3 * time.Second,
				MinTimeout: 2 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return resource.RetryableError(fmt.Errorf("error on waiting for stopping instance when deleting %q, %s", d.Id(), err))
			}
		}

		if _, err := conn.TerminateUHostInstance(deleReq); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting instance %q, %s", d.Id(), err))
		}

		if _, err := client.describeInstanceById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error on reading instance when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified instance %q has not been deleted due to unknown error", d.Id()))
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

func diffValidateDefaultSecurityGroup(old, new, meta interface{}) error {
	client := meta.(*UCloudClient)

	// check default firewall is exists when no firewall is specified
	if len(new.(string)) == 0 {
		return client.checkDefaultFirewall()
	}
	return nil
}

func diffValidateInstanceDataDiskSize(old, new, meta interface{}) error {

	if new.(int) < old.(int) {
		return fmt.Errorf("reduce data disk size is not supported, "+
			"new value %d should be larger than the old value %d", new.(int), old.(int))
	}
	return nil
}

func diffValidateInstanceBootDiskSize(old, new, meta interface{}) error {

	if new.(int) < old.(int) {
		return fmt.Errorf("reduce boot disk size is not supported, "+
			"new value %d by user set should be larger than the old value %d allocated by the system", new.(int), old.(int))
	}
	return nil
}

func diffValidateBootDiskTypeWithDataDiskType(diff *schema.ResourceDiff, v interface{}) error {
	var dataDiskType string
	var bootDiskType string

	if v, ok := diff.GetOk("boot_disk_type"); ok {
		bootDiskType = v.(string)
	} else {
		bootDiskType = "local_normal"
	}

	if _, ok := diff.GetOk("data_disk_size"); ok {
		if v, ok := diff.GetOk("data_disk_type"); ok {
			dataDiskType = v.(string)
		} else {
			dataDiskType = "local_normal"
		}
	}
	if checkStringIn(bootDiskType, []string{"cloud_normal", "cloud_ssd"}) == nil && checkStringIn(dataDiskType, []string{"local_normal", "local_ssd"}) == nil {
		return fmt.Errorf("the instance cannot have local data disk, When the %q is %q", "boot_disk_type", bootDiskType)
	}
	return nil
}

func diffValidateInstanceTypeWithZone(diff *schema.ResourceDiff, v interface{}) error {
	t, _ := parseInstanceType(diff.Get("instance_type").(string))
	zone := diff.Get("availability_zone").(string)

	if t.HostType == "o" && zone != "cn-bj2-05" {
		return fmt.Errorf("the outstanding type about %q only be supported in %q, got %q", "instance_type", "cn-bj2-05", zone)
	}

	return nil
}
