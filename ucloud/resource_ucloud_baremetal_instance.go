package ucloud

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/uphost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudBareMetalInstance() *schema.Resource {
	return &schema.Resource{
		Create:        resourceUCloudBareMetalInstanceCreate,
		Read:          resourceUCloudBareMetalInstanceRead,
		Update:        resourceUCloudBareMetalInstanceUpdate,
		Delete:        resourceUCloudBareMetalInstanceDelete,
		CustomizeDiff: bareMetalInstanceCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"allow_stopping_for_update": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allow_stopping_for_resizing": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"delete_disks_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"delete_eips_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"root_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validateUcloudInstanceRootPassword,
			},
			"boot_disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"boot_disk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(20, 500),
			},
			"boot_disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"cloud_rssd",
				}, false),
			},
			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "day",
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"day",
				}, false),
			},
			"duration": {
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"remark": {
				Type:     schema.TypeString,
				Optional: true,
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
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"data_disks": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(20, 8000),
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"cloud_rssd",
							}, false),
						},
					},
				},
			},
			"network_interface": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eip_bandwidth": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 800),
						},
						"eip_internet_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"international",
								"bgp",
							}, false),
						},
						"eip_charge_mode": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"traffic",
								"bandwidth",
							}, false),
						},
					},
				},
			},
			"raid_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"raid1",
					"raid0",
					"raid10",
					"raid5",
					"no_raid",
				}, false),
			},
		},
	}
}
func resourceUCloudBareMetalInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	req := conn.NewCreatePHostRequest()
	req.SetEncoder(request.NewJSONEncoder(conn.GetConfig(), conn.GetCredential()))
	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.ImageId = ucloud.String(d.Get("image_id").(string))
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("root_password").(string))))
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Quantity = ucloud.String(strconv.Itoa(d.Get("duration").(int)))
	req.Name = ucloud.String(d.Get("name").(string))
	req.Tag = ucloud.String(d.Get("tag").(string))
	req.Remark = ucloud.String(d.Get("remark").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	if privateIp, ok := d.GetOk("private_ip"); ok {
		req.VpcIp = ucloud.String(privateIp.(string))
	}
	if securityGroup, ok := d.GetOk("security_group"); ok {
		firewall, err := client.describeFirewallById(securityGroup.(string))
		if err != nil {
			return fmt.Errorf("fail to retrieve firewall: %v", err)
		}
		req.SecurityGroupId = ucloud.String(firewall.GroupId)
	}

	// Get instance type
	instanceType := d.Get("instance_type").(string)
	var isCloudDiskInstance bool
	var isLocalDiskInstance bool

	// Create a request object for DescribePHostMachineType
	describePHostReq := conn.NewDescribePHostMachineTypeRequest()

	// Call DescribePHostMachineType API to get valid types for local disk instance
	localDiskTypesResp, err := conn.DescribePHostMachineType(describePHostReq)
	if err != nil {
		return fmt.Errorf("error on getting local disk types, %s", err)
	}
	for _, machineType := range localDiskTypesResp.MachineTypes {
		if machineType.Type == instanceType {
			for _, cluster := range machineType.Clusters {
				if cluster.StockStatus != "SoldOut" {
					req.Cluster = ucloud.String(cluster.Name)
					break
				}
			}
			isLocalDiskInstance = true
		}
	}

	// Create a request object for DescribeBareMetalMachineType
	describeBareMetalReq := conn.NewDescribeBaremetalMachineTypeRequest()

	// Call DescribeBareMetalMachineType API to get valid types for cloud disk instance
	cloudDiskTypesResp, err := conn.DescribeBaremetalMachineType(describeBareMetalReq)
	if err != nil {
		return fmt.Errorf("error on getting cloud disk types, %s", err)
	}
	for _, machineType := range cloudDiskTypesResp.MachineTypes {
		if machineType.Type == instanceType {
			for _, cluster := range machineType.Clusters {
				if cluster.StockStatus != "SoldOut" {
					req.Cluster = ucloud.String(cluster.Name)
					break
				}
			}
			isCloudDiskInstance = true
		}
	}
	if req.Cluster == nil {
		return fmt.Errorf("resource of %v is not available", instanceType)
	}
	// Check if instance type is a valid type
	if isLocalDiskInstance {
		req.Type = ucloud.String(instanceType)
		if _, ok := d.GetOk("raid_type"); !ok {
			return fmt.Errorf("raid_type is required for local disk instance")
		}
		req.Raid = ucloud.String(raidTypeCvt.unconvert(d.Get("raid_type").(string)))
	} else if isCloudDiskInstance {
		req.Type = ucloud.String(instanceType)
		if _, ok := d.GetOk("boot_disk_size"); !ok {
			return fmt.Errorf("boot_disk_size is required for cloud disk instance")
		}
		if _, ok := d.GetOk("boot_disk_type"); !ok {
			return fmt.Errorf("boot_disk_type is required for cloud disk instance")
		}
		bootDisk := uphost.CreatePHostParamDisks{
			Size:   ucloud.Int(d.Get("boot_disk_size").(int)),
			Type:   ucloud.String(d.Get("boot_disk_type").(string)),
			IsBoot: ucloud.String("True"),
		}
		req.Disks = append(req.Disks, bootDisk)
		if _, ok := d.GetOk("data_disks"); ok {
			disks := d.Get("data_disks").([]interface{})
			for _, disk := range disks {
				dataDisk := uphost.CreatePHostParamDisks{
					Size:   ucloud.Int(disk.(map[string]interface{})["size"].(int)),
					Type:   ucloud.String(disk.(map[string]interface{})["type"].(string)),
					IsBoot: ucloud.String("False"),
				}
				req.Disks = append(req.Disks, dataDisk)
			}
		}
	} else {
		return fmt.Errorf("invalid instance type: %s", instanceType)
	}

	if val, ok := d.GetOk("network_interface"); ok {
		interfaces := val.([]interface{})
		for _, iface := range interfaces {
			ifaceMap := iface.(map[string]interface{})
			networkInterface := uphost.CreatePHostParamNetworkInterface{
				EIP: &uphost.CreatePHostParamNetworkInterfaceEIP{
					CouponId:     ucloud.String(""),
					Bandwidth:    ucloud.String(strconv.Itoa(ifaceMap["eip_bandwidth"].(int))),
					PayMode:      ucloud.String(upperCamelCvt.unconvert(ifaceMap["eip_charge_mode"].(string))),
					OperatorName: ucloud.String(upperCamelCvt.unconvert(ifaceMap["eip_internet_type"].(string))),
				},
			}
			req.NetworkInterface = append(req.NetworkInterface, networkInterface)
		}
	}

	resp, err := conn.CreatePHost(req)

	if err != nil {
		return fmt.Errorf("error on creating bare metal instance, %s", err)
	}

	d.SetId(resp.PHostId[0])

	// Wait for instance to be in "Running" state before returning
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Starting"},
		Target:     []string{"Running"},
		Refresh:    bareMetalInstanceStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for bare metal instance %s to be running, %s", d.Id(), err)
	}

	return resourceUCloudBareMetalInstanceRead(d, meta)
}

func resourceUCloudBareMetalInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	req := conn.NewDescribePHostRequest()
	req.PHostId = []string{d.Id()}

	resp, err := conn.DescribePHost(req)
	if err != nil {
		return fmt.Errorf("error on reading bare metal instance %s, %s", d.Id(), err)
	}
	if len(resp.PHostSet) == 0 {
		return newNotFoundError("resource cannot be found")
	}
	networkInterfaces := make([]map[string]interface{}, 0)
	instance := resp.PHostSet[0]
	for _, item := range instance.IPSet {
		if item.OperatorName == "Private" {
			d.Set("vpc_id", item.VPCId)
			d.Set("subnet_id", item.SubnetId)
			d.Set("private_ip", item.IPAddr)
			break
		} else {
			eipPayModeReq := client.unetconn.NewGetEIPPayModeRequest()
			eipPayModeReq.EIPId = []string{item.IPId}
			eipPayModeResp, err := client.unetconn.GetEIPPayMode(eipPayModeReq)
			if err != nil {
				return fmt.Errorf("error on reading eip_charge_mode when reading instance %q, %s", d.Id(), err)
			}
			if len(eipPayModeResp.EIPPayMode) == 0 {
				return fmt.Errorf("fail to get eip_charge_mode when reading instance %q", d.Id())
			}

			networkInterfaces = append(networkInterfaces, map[string]interface{}{
				"eip_bandwidth":     item.Bandwidth,
				"eip_internet_type": upperCvt.convert(item.OperatorName),
				"eip_charge_mode":   upperCamelCvt.convert(eipPayModeResp.EIPPayMode[0].EIPPayMode),
			})
		}
	}
	dataDisks := make([]map[string]interface{}, 0)
	for _, item := range instance.DiskSet {
		diskType := upperCvt.convert(item.Type)
		if item.IsBoot == "True" {
			d.Set("boot_disk_size", item.Space)
			d.Set("boot_disk_type", diskType)
			d.Set("boot_disk_id", item.DiskId)
		} else {
			dataDisks = append(dataDisks, map[string]interface{}{
				"id":          item.DiskId,
				"device_name": item.Drive,
				"size":        item.Space,
				"type":        diskType,
			})
		}
	}
	// Set the properties of the instance
	d.Set("availability_zone", instance.Zone)
	d.Set("charge_type", ucloud.String(upperCamelCvt.convert(instance.ChargeType)))
	d.Set("name", instance.Name)
	d.Set("tag", instance.Tag)
	d.Set("remark", instance.Remark)

	raidType, err := client.getRaidTypeById(d.Id())
	if err != nil {
		return fmt.Errorf("error on reading raid type when reading instance %q, %s", d.Id(), err)
	}
	if raidType != "" {
		d.Set("raid_type", raidTypeCvt.convert(raidType))
	}
	sgSet, err := client.describeFirewallByIdAndType(d.Id(), eipResourceTypeUPHost)
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error on reading security group when reading instance %q, %s", d.Id(), err)
	}

	d.Set("security_group", sgSet.FWId)

	if _, ok := d.GetOk("data_disks"); ok {
		d.Set("data_disks", dataDisks)
	}
	if _, ok := d.GetOk("network_interface"); ok {
		d.Set("network_interface", networkInterfaces)
	}
	return nil
}

func resourceUCloudBareMetalInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	updateFunc := func(updateAction func() error) error {
		err := stopUpdateStartBareMetalInstance(d, meta, updateAction)
		if err != nil {
			return err
		}
		return nil
	}
	zone := ucloud.String(d.Get("availability_zone").(string))
	d.Partial(true)
	if d.HasChange("root_password") {
		if _, ok := d.GetOk("allow_stopping_for_update"); ok {
			err := updateFunc(func() error {
				resetReq := conn.NewResetPHostPasswordRequest()
				resetReq.Zone = zone
				resetReq.PHostId = ucloud.String(d.Id())
				resetReq.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("root_password").(string))))
				if _, err := conn.ResetPHostPassword(resetReq); err != nil {
					return fmt.Errorf("error on resetting password, %s", err)
				}
				return nil
			})

			if err != nil {
				return err
			}
			d.SetPartial("root_password")
		} else {
			return errors.New("allow_stopping_for_update must be true, when root_password needs to be updated")
		}
	}

	if d.HasChange("image_id") {
		err := updateFunc(func() error {
			reinstallReq := conn.NewReinstallPHostRequest()
			reinstallReq.Zone = zone
			reinstallReq.PHostId = ucloud.String(d.Id())
			reinstallReq.ImageId = ucloud.String(d.Get("image_id").(string))

			if _, err := conn.ReinstallPHost(reinstallReq); err != nil {
				return fmt.Errorf("error on reinstalling instance with new image_id, %s", err)
			}
			return nil
		})

		if err != nil {
			return err
		}
		d.SetPartial("image_id")
	}

	if d.HasChanges("boot_disk_size") {
		resizeRequests := make([]*uphost.ResizePHostAttachedDiskRequest, 0)
		if d.HasChange("boot_disk_size") {
			resizeReq := conn.NewResizePHostAttachedDiskRequest()
			resizeReq.Zone = zone
			resizeReq.PHostId = ucloud.String(d.Id())
			resizeReq.DiskSpace = ucloud.Int(d.Get("boot_disk_size").(int))
			resizeReq.UDiskId = ucloud.String(d.Get("boot_disk_id").(string))
			resizeRequests = append(resizeRequests, resizeReq)
		}

		if _, ok := d.GetOk("allow_stopping_for_resizing"); ok {
			err := updateFunc(func() error {
				for _, req := range resizeRequests {
					if _, err := conn.ResizePHostAttachedDisk(req); err != nil {
						return fmt.Errorf("error on resizing disk %s", err)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			for _, req := range resizeRequests {
				if _, err := conn.ResizePHostAttachedDisk(req); err != nil {
					return fmt.Errorf("error on resizing disk %s", err)
				}
			}
		}
		d.SetPartial("boot_disk_size")
	}

	if d.HasChange("name") || d.HasChange("remark") || d.HasChange("tag") {
		modifyInfoReq := conn.NewModifyPHostInfoRequest()
		modifyInfoReq.PHostId = ucloud.String(d.Id())
		modifyInfoReq.Name = ucloud.String(d.Get("name").(string))
		modifyInfoReq.Remark = ucloud.String(d.Get("remark").(string))
		modifyInfoReq.Tag = ucloud.String(d.Get("tag").(string))
		modifyInfoReq.Zone = zone
		if _, err := conn.ModifyPHostInfo(modifyInfoReq); err != nil {
			return fmt.Errorf("error on updating name, remark or tag, %s", err)
		}
		d.SetPartial("name")
		d.SetPartial("remark")
		d.SetPartial("tag")
	}
	if d.HasChange("security_group") {
		conn := client.unetconn
		req := conn.NewGrantFirewallRequest()
		req.FWId = ucloud.String(d.Get("security_group").(string))
		req.ResourceType = ucloud.String(eipResourceTypeUPHost)
		req.ResourceId = ucloud.String(d.Id())
		req.Zone = zone
		_, err := conn.GrantFirewall(req)
		if err != nil {
			return fmt.Errorf("error on %s to instance %q, %s", "GrantFirewall", d.Id(), err)
		}

		d.SetPartial("security_group")
	}
	d.Partial(false)
	return resourceUCloudBareMetalInstanceRead(d, meta)
}

func resourceUCloudBareMetalInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	deleReq := conn.NewTerminatePHostRequest()
	deleReq.PHostId = ucloud.String(d.Id())
	_, releaseUDisk := d.GetOk("delete_disks_with_instance")
	_, releaseEIP := d.GetOk("delete_eips_with_instance")

	deleReq.ReleaseUDisk = ucloud.Bool(releaseUDisk)
	deleReq.ReleaseEIP = ucloud.Bool(releaseEIP)

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		instance, err := client.describeBareMetalInstanceById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading instance before deleting %q, %s", d.Id(), err))
		}
		stopReq := conn.NewPoweroffPHostRequest()
		stopReq.PHostId = ucloud.String(d.Id())
		if !isStringIn(instance.PMStatus, []string{statusStopped, instanceStatusInstallFail, instanceStatusResizeFail}) {
			if _, err := conn.PoweroffPHost(stopReq); err != nil {
				return resource.RetryableError(fmt.Errorf("error on stopping instance when deleting %q, %s", d.Id(), err))
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{uphostStatusStopping},
				Target:     []string{statusStopped},
				Refresh:    bareMetalInstanceStateRefreshFunc(client, d.Id()),
				Timeout:    5 * time.Minute,
				Delay:      3 * time.Second,
				MinTimeout: 2 * time.Second,
			}

			if _, err = stateConf.WaitForState(); err != nil {
				return resource.RetryableError(fmt.Errorf("error on waiting for stopping instance when deleting %q, %s", d.Id(), err))
			}
		}

		if _, err := conn.TerminatePHost(deleReq); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting instance %q, %s", d.Id(), err))
		}

		if _, err := client.describeBareMetalInstanceById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error on reading instance when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified instance %q has not been deleted due to unknown error", d.Id()))
	})
	return nil
}

func bareMetalInstanceStateRefreshFunc(client *UCloudClient, instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.describeBareMetalInstanceById(instanceId)
		if err != nil {
			if isNotFoundError(err) {
				return nil, "", fmt.Errorf("instance not found")
			}
			return nil, "", err
		}

		// Assuming that State is a field of PHostSet
		// Adjust this according to the actual structure of PHostSet
		return resp, resp.PMStatus, nil
	}
}

func validateUcloudInstanceRootPassword(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 8 || len(value) > 30 {
		errors = append(errors, fmt.Errorf(
			"%q must be between 8 and 30 characters", k))
	}

	if !regexp.MustCompile(`^[A-Za-z0-9_]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q can only contain alphanumeric characters and underscores", k))
	}

	return
}

func stopUpdateStartBareMetalInstance(d *schema.ResourceData, meta interface{}, updateFunc func() error) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	stopReq := conn.NewTerminatePHostRequest()
	stopReq.PHostId = ucloud.String(d.Id())

	if _, err := conn.TerminatePHost(stopReq); err != nil {
		return fmt.Errorf("error on stopping instance when updating, %s", err)
	}

	// Wait for instance to be in "Stopped" state before updating
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Running", "Stopping"},
		Target:     []string{"Stopped"},
		Refresh:    bareMetalInstanceStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for instance to be stopped when updating, %s", err)
	}

	// Perform the update operation
	if err := updateFunc(); err != nil {
		return err
	}

	startReq := conn.NewStartPHostRequest()
	startReq.PHostId = ucloud.String(d.Id())

	if _, err := conn.StartPHost(startReq); err != nil {
		return fmt.Errorf("error on starting instance after updating, %s", err)
	}

	// Wait for instance to be in "Running" state after starting
	stateConf = &resource.StateChangeConf{
		Pending:    []string{"Stopped", "Starting"},
		Target:     []string{"Running"},
		Refresh:    bareMetalInstanceStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for instance to be started after updating, %s", err)
	}

	return nil
}

func bareMetalInstanceCustomizeDiff(diff *schema.ResourceDiff, v interface{}) error {
	client := v.(*UCloudClient)
	conn := client.uphostconn

	instanceType := diff.Get("instance_type").(string)
	baremetalMachineTypeRequest := conn.NewDescribeBaremetalMachineTypeRequest()
	cloudDiskTypesResp, err := conn.DescribeBaremetalMachineType(baremetalMachineTypeRequest)
	if err != nil {
		return fmt.Errorf("error on getting cloud disk types, %s", err)
	}
	phostMachineTypeRequest := conn.NewDescribePHostMachineTypeRequest()
	localDiskTypesResp, err := conn.DescribePHostMachineType(phostMachineTypeRequest)
	if err != nil {
		return fmt.Errorf("error on getting local disk types, %s", err)
	}
	cloudDiskTypes := []string{}
	for _, machineType := range cloudDiskTypesResp.MachineTypes {
		cloudDiskTypes = append(cloudDiskTypes, machineType.Type)
	}
	localDiskTypes := []string{}
	for _, machineType := range localDiskTypesResp.MachineTypes {
		localDiskTypes = append(localDiskTypes, machineType.Type)
	}
	if isStringIn(instanceType, cloudDiskTypes) {
		if _, ok := diff.GetOk("boot_disk_size"); !ok {
			return fmt.Errorf("boot_disk_size is required for cloud disk instance")
		}
		if _, ok := diff.GetOk("boot_disk_type"); !ok {
			return fmt.Errorf("boot_disk_type is required for cloud disk instance")
		}
		if _, ok := diff.GetOk("data_disks"); !ok {
			return fmt.Errorf("data_disks is required for cloud disk instance")
		}
		if _, ok := diff.GetOk("raid_type"); ok {
			return fmt.Errorf("raid_type should not be set for cloud disk instance")
		}
	} else if isStringIn(instanceType, localDiskTypes) {
		if _, ok := diff.GetOk("raid_type"); !ok {
			return fmt.Errorf("raid_type is required for local disk instance")
		}
		if _, ok := diff.GetOk("boot_disk_size"); ok {
			return fmt.Errorf("boot_disk_size should not be set for local disk instance")
		}
		if _, ok := diff.GetOk("boot_disk_type"); ok {
			return fmt.Errorf("boot_disk_type should not be set for local disk instance")
		}
		if _, ok := diff.GetOk("data_disks"); ok {
			return fmt.Errorf("data_disks should not be set for local disk instance")
		}
	} else {
		return fmt.Errorf("invalid instance type: %s", instanceType)
	}

	return nil
}
