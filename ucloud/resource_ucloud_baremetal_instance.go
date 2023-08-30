package ucloud

import (
	"fmt"
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
		Create: resourceUCloudBareMetalInstanceCreate,
		Read:   resourceUCloudBareMetalInstanceRead,
		Update: resourceUCloudBareMetalInstanceUpdate,
		Delete: resourceUCloudBareMetalInstanceDelete,
		CustomizeDiff: func(diff *schema.ResourceDiff, v interface{}) error {
			client := v.(*UCloudClient)
			conn := client.uphostconn

			instanceType := diff.Get("instance_type").(string)
			cloudDiskTypes, err := conn.DescribeBareMetalMachineType()
			if err != nil {
				return fmt.Errorf("error on getting cloud disk types, %s", err)
			}

			localDiskTypes, err := conn.DescribePHostMachineType()
			if err != nil {
				return fmt.Errorf("error on getting local disk types, %s", err)
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
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
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
			"root_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validateUcloudInstanceRootPassword,
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
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"day",
				}, false),
			},
			"duration": {
				Type:     schema.TypeInt,
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
			"security_group": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 63),
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"data_disks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
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
			"delete_disks_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"network_interface": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
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

	req.Zone = ucloud.String(d.Get("availability_zone").(string))
	req.ImageId = ucloud.String(d.Get("image_id").(string))
	req.Password = ucloud.String(d.Get("root_password").(string))
	req.ChargeType = ucloud.String(d.Get("charge_type").(string))
	req.Quantity = ucloud.String(strconv.Itoa(d.Get("duration").(int)))
	req.Name = ucloud.String(d.Get("name").(string))
	req.Tag = ucloud.String(d.Get("tag").(string))
	req.Remark = ucloud.String(d.Get("remark").(string))
	req.SecurityGroupId = ucloud.String(d.Get("security_group").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))
	req.VpcIp = ucloud.String(d.Get("private_ip").(string))

	// Get instance type
	instanceType := d.Get("instance_type").(string)

	// Create a request object for DescribePHostMachineType
	describePHostReq := conn.NewDescribePHostMachineTypeRequest()
	describePHostReq.Zone = ucloud.String(d.Get("availability_zone").(string))

	// Call DescribePHostMachineType API to get valid types for local disk instance
	localDiskTypesResp, err := conn.DescribePHostMachineType(describePHostReq)
	if err != nil {
		return fmt.Errorf("error on getting local disk types, %s", err)
	}
	localDiskTypes := []string{}
	for _, machineType := range localDiskTypesResp.MachineTypes {
		localDiskTypes = append(localDiskTypes, machineType.Type)
	}

	// Create a request object for DescribeBareMetalMachineType
	describeBareMetalReq := conn.NewDescribeBaremetalMachineTypeRequest()
	describeBareMetalReq.Zone = ucloud.String(d.Get("availability_zone").(string))

	// Call DescribeBareMetalMachineType API to get valid types for cloud disk instance
	cloudDiskTypesResp, err := conn.DescribeBaremetalMachineType(describeBareMetalReq)
	if err != nil {
		return fmt.Errorf("error on getting cloud disk types, %s", err)
	}
	cloudDiskTypes := []string{}
	for _, machineType := range cloudDiskTypesResp.MachineTypes {
		cloudDiskTypes = append(cloudDiskTypes, machineType.Type)
	}

	// Check if instance type is a valid type
	if isStringIn(instanceType, localDiskTypes) {
		req.Type = ucloud.String(instanceType)
		if _, ok := d.GetOk("raid_type"); !ok {
			return fmt.Errorf("raid_type is required for local disk instance")
		}
		req.Raid = ucloud.String(d.Get("raid_type").(string))
	} else if isStringIn(instanceType, cloudDiskTypes) {
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
					Bandwidth:    ucloud.String(strconv.Itoa(ifaceMap["eip_bandwidth"].(int))),
					PayMode:      ucloud.String(ifaceMap["eip_charge_mode"].(string)),
					OperatorName: ucloud.String(ifaceMap["eip_internet_type"].(string)),
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
		Pending:    []string{"Pending"},
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
	// Set the properties of the instance
	d.Set("availability_zone", resp.PHostSet[0].Zone)
	d.Set("charge_type", resp.PHostSet[0].ChargeType)
	d.Set("name", resp.PHostSet[0].Name)
	d.Set("tag", resp.PHostSet[0].Tag)
	d.Set("remark", resp.PHostSet[0].Remark)
	d.Set("vpc_id", resp.PHostSet[0])
	d.Set("subnet_id", resp.PHostSet[0].SubnetId)
	d.Set("private_ip", resp.PHostSet[0].PrivateIp)
	d.Set("instance_type", resp.PHosts[0].InstanceType)
	d.Set("raid_type", resp.PHosts[0].Raid)
	d.Set("boot_disk_size", resp.PHosts[0].BootDiskSpace)
	d.Set("boot_disk_type", resp.PHosts[0].BootDiskType)
	d.Set("data_disks", resp.PHosts[0].DataDisks)
	d.Set("network_interface", resp.PHosts[0].NetworkInterface)

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

	if d.HasChange("root_password") {
		err := updateFunc(func() error {
			resetReq := conn.NewResetPHostPasswordRequest()
			resetReq.PHostId = ucloud.String(d.Id())
			resetReq.Password = ucloud.String(d.Get("root_password").(string))

			if _, err := conn.ResetPHostPassword(resetReq); err != nil {
				return fmt.Errorf("error on resetting password, %s", err)
			}
			return nil
		})

		if err != nil {
			return err
		}
	}

	if d.HasChange("image_id") {
		err := updateFunc(func() error {
			reinstallReq := conn.NewReinstallPHostRequest()
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
	}

	if d.HasChange("name") || d.HasChange("remark") || d.HasChange("tag") {
		modifyInfoReq := conn.NewModifyPHostInfoRequest()
		modifyInfoReq.PHostId = ucloud.String(d.Id())
		modifyInfoReq.Name = ucloud.String(d.Get("name").(string))
		modifyInfoReq.Remark = ucloud.String(d.Get("remark").(string))
		modifyInfoReq.Tag = ucloud.String(d.Get("tag").(string))

		if _, err := conn.ModifyPHostInfo(modifyInfoReq); err != nil {
			return fmt.Errorf("error on updating name, remark or tag, %s", err)
		}
	}

	return resourceUCloudBareMetalInstanceRead(d, meta)
}

func resourceUCloudBareMetalInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uphostconn

	req := conn.NewDeletePHostInstanceRequest()
	req.PHostId = ucloud.String(d.Id())

	_, err := conn.DeletePHostInstance(req)
	if err != nil {
		return fmt.Errorf("error on deleting bare metal instance %s, %s", d.Id(), err)
	}

	// Wait for instance to be in "Terminated" state before returning
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Running", "Stopping"},
		Target:     []string{"Terminated"},
		Refresh:    bareMetalInstanceStateRefreshFunc(client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for bare metal instance %s to be terminated, %s", d.Id(), err)
	}

	return nil
}

func bareMetalInstanceStateRefreshFunc(client *UCloudClient, instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		req := uphost.DescribePHostRequest()
		req.PHostIds = []string{instanceId}

		resp, err := client.uphostconn.DescribePHost(req)
		if err != nil {
			if isNotFoundError(err) {
				return nil, "", fmt.Errorf("instance not found")
			}
			return nil, "", err
		}

		if len(resp.PHostSet) == 0 {
			return nil, "", fmt.Errorf("instance not found")
		}

		// Assuming that State is a field of PHostSet
		// Adjust this according to the actual structure of PHostSet
		return resp.PHostSet[0], resp.PHostSet[0].State, nil
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

	stopReq := conn.NewStopPHostInstanceRequest()
	stopReq.PHostId = ucloud.String(d.Id())

	if _, err := conn.StopPHostInstance(stopReq); err != nil {
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

	startReq := conn.NewStartPHostInstanceRequest()
	startReq.PHostId = ucloud.String(d.Id())

	if _, err := conn.StartPHostInstance(startReq); err != nil {
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