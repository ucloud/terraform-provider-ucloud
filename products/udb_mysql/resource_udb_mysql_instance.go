package udb_mysql

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceUCloudMySQLInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudMySQLInstanceCreate,
		Read:   resourceUCloudMySQLInstanceRead,
		Update: resourceUCloudMySQLInstanceUpdate,
		Delete: resourceUCloudMySQLInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(6, 63),
			},

			"db_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "mysql-8.0",
				ValidateFunc: validation.StringInSlice(dbVersionList, false),
			},

			"machine_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(dbMachineTypeList, false),
			},

			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(8, 30),
			},

			"instance_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "HA",
				ValidateFunc: validation.StringInSlice([]string{"Normal", "HA"}, false),
			},

			"disk_space": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(20, 32000),
			},

			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      3306,
				ValidateFunc: validation.IntBetween(3306, 65535),
			},

			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "Month",
				ValidateFunc: validation.StringInSlice([]string{"Month", "Year", "Dynamic"}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 36),
			},

			"param_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

			"tag": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"memory": {
				Type:     schema.TypeInt,
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

			"modify_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUCloudMySQLInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating mysql instance, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	password := d.Get("password").(string)
	if password == "" {
		password = fmt.Sprintf("%s%s%s",
			acctest.RandStringFromCharSet(5, defaultPasswordStr),
			acctest.RandStringFromCharSet(1, defaultPasswordSpe),
			acctest.RandStringFromCharSet(5, defaultPasswordNum))
	}

	name := d.Get("name").(string)
	if name == "" {
		name = resource.PrefixedUniqueId("tf-mysql-instance-")
	}

	// resolve parameter group: user supplied or the default template
	var paramGroupID int
	if v, ok := d.GetOk("param_group_id"); ok {
		paramGroupID, err = strconv.Atoi(v.(string))
		if err != nil {
			return fmt.Errorf("error on setting param_group_id %q when creating mysql instance, %s", v.(string), err)
		}
	} else {
		paramGroupID, err = getDefaultParamGroupID(client, zone, d.Get("db_version").(string))
		if err != nil {
			return err
		}
	}

	payload := basePayload(client, zone)
	payload["Name"] = name
	payload["AdminPassword"] = password
	payload["DBTypeId"] = d.Get("db_version").(string)
	payload["DiskSpace"] = d.Get("disk_space").(int)
	payload["ParamGroupId"] = paramGroupID
	payload["MachineType"] = d.Get("machine_type").(string)
	payload["StorageClass"] = storageClassNvme
	payload["SpecificationClass"] = specificationClassNvme
	payload["InstanceMode"] = d.Get("instance_mode").(string)
	payload["ChargeType"] = d.Get("charge_type").(string)
	payload["Quantity"] = d.Get("duration").(int)
	payload["Port"] = d.Get("port").(int)

	if v, ok := d.GetOk("vpc_id"); ok {
		payload["VPCId"] = v.(string)
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		payload["SubnetId"] = v.(string)
	}
	if v, ok := d.GetOk("tag"); ok {
		payload["Tag"] = v.(string)
	}

	resp, err := invoke(client, "CreateUDBMySQLInstance", payload)
	if err != nil {
		return fmt.Errorf("error on creating mysql instance, %s", err)
	}

	dbId := payloadString(resp, "DBId")
	if dbId == "" {
		return fmt.Errorf("empty DBId in create response")
	}
	d.SetId(dbId)
	if _, ok := d.GetOk("password"); !ok {
		d.Set("password", password)
	}
	if _, ok := d.GetOk("name"); !ok {
		d.Set("name", name)
	}

	// wait until the instance is running
	stateConf := resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{mysqlStateRunning},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
		Refresh:    mysqlInstanceStateRefreshFunc(client, dbId, zone, []string{mysqlStateRunning}),
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for mysql instance %q complete creating, %s", dbId, err)
	}

	return resourceUCloudMySQLInstanceRead(d, meta)
}

func resourceUCloudMySQLInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading mysql instance, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	db, err := describeDBInstanceByID(client, d.Id(), zone)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading mysql instance %q, %s", d.Id(), err)
	}

	d.Set("name", payloadString(db, "Name"))
	d.Set("db_version", payloadString(db, "DBTypeId"))
	d.Set("instance_mode", normalizeInstanceMode(payloadString(db, "InstanceMode")))
	d.Set("port", payloadInt(db, "Port"))
	d.Set("private_ip", payloadString(db, "VirtualIP"))
	d.Set("status", payloadString(db, "State"))
	d.Set("cpu", payloadInt(db, "CPU"))
	d.Set("memory", payloadInt(db, "MemoryLimit")/1000)
	d.Set("disk_space", payloadInt(db, "DiskSpace"))
	d.Set("charge_type", payloadString(db, "ChargeType"))
	d.Set("vpc_id", payloadString(db, "VPCId"))
	d.Set("subnet_id", payloadString(db, "SubnetId"))
	d.Set("tag", payloadString(db, "Tag"))
	d.Set("param_group_id", strconv.Itoa(payloadInt(db, "ParamGroupId")))
	d.Set("create_time", timestampToString(payloadInt(db, "CreateTime")))
	d.Set("expire_time", timestampToString(payloadInt(db, "ExpiredTime")))
	d.Set("modify_time", timestampToString(payloadInt(db, "ModifyTime")))

	return nil
}

func resourceUCloudMySQLInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating mysql instance, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	if d.HasChange("machine_type") || d.HasChange("disk_space") {
		payload := basePayload(client, zone)
		payload["DBId"] = d.Id()
		// for nvme single instance, resource_id equals DBId
		payload["resource_id"] = d.Id()
		payload["MachineType"] = d.Get("machine_type").(string)
		payload["SpecificationType"] = 1
		payload["DiskSpace"] = d.Get("disk_space").(int)
		payload["StorageClass"] = storageClassNvme
		payload["SpecificationClass"] = specificationClassNvme

		if _, err := invoke(client, "ResizeUDBInstance", payload); err != nil {
			return fmt.Errorf("error on resizing mysql instance %q, %s", d.Id(), err)
		}

		stateConf := resource.StateChangeConf{
			Pending:    []string{statusPending},
			Target:     []string{mysqlStateRunning},
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      3 * time.Second,
			MinTimeout: 2 * time.Second,
			Refresh:    mysqlInstanceStateRefreshFunc(client, d.Id(), zone, []string{mysqlStateRunning}),
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for resizing mysql instance %q, %s", d.Id(), err)
		}
	}

	return resourceUCloudMySQLInstanceRead(d, meta)
}

func resourceUCloudMySQLInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting mysql instance, %s", err)
	}
	zone := d.Get("availability_zone").(string)
	dbId := d.Id()

	return resource.Retry(10*time.Minute, func() *resource.RetryError {
		db, err := describeDBInstanceByID(client, dbId, zone)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		state := payloadString(db, "State")
		if !isStringIn(state, []string{mysqlStateShutoff, mysqlStateFail, mysqlStateDeleteFail, mysqlStateRecoverFail}) {
			stopPayload := basePayload(client, zone)
			stopPayload["DBId"] = dbId
			if _, err := invoke(client, "StopUDBInstance", stopPayload); err != nil {
				return resource.RetryableError(fmt.Errorf("error on stopping mysql instance when deleting %q, %s", dbId, err))
			}

			stateConf := resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{mysqlStateShutoff},
				Timeout:    d.Timeout(schema.TimeoutDelete),
				Delay:      3 * time.Second,
				MinTimeout: 2 * time.Second,
				Refresh:    mysqlInstanceStateRefreshFunc(client, dbId, zone, []string{mysqlStateShutoff}),
			}
			if _, err := stateConf.WaitForState(); err != nil {
				return resource.RetryableError(fmt.Errorf("error on waiting for stopping mysql instance when deleting %q, %s", dbId, err))
			}
		}

		deletePayload := basePayload(client, zone)
		deletePayload["DBId"] = dbId
		if _, err := invoke(client, "DeleteUDBInstance", deletePayload); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting mysql instance %q, %s", dbId, err))
		}

		if _, err := describeDBInstanceByID(client, dbId, zone); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading mysql instance when deleting %q, %s", dbId, err))
		}

		return resource.RetryableError(fmt.Errorf("the specified mysql instance %q has not been deleted due to unknown error", dbId))
	})
}
