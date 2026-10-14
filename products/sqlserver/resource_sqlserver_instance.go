package sqlserver

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceUCloudSQLServerInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudSQLServerInstanceCreate,
		Read:   resourceUCloudSQLServerInstanceRead,
		Delete: resourceUCloudSQLServerInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
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
				Default:      "sqlserver-2019",
				ValidateFunc: validation.StringInSlice(dbVersionList, false),
			},

			"cpu": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(2, 64),
			},

			"memory": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(2000, 320000),
			},

			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(8, 30),
			},

			"disk_space": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(20, 32000),
			},

			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      1433,
				ValidateFunc: validation.IntBetween(1433, 65535),
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

func resourceUCloudSQLServerInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating sqlserver instance, %s", err)
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
		name = resource.PrefixedUniqueId("tf-sqlserver-instance-")
	}

	payload := basePayload(client, zone)
	payload["Name"] = name
	payload["AdminPassword"] = password
	payload["DBTypeId"] = d.Get("db_version").(string)
	payload["Port"] = d.Get("port").(int)
	payload["DiskSpace"] = d.Get("disk_space").(int)
	payload["CPU"] = d.Get("cpu").(int)
	payload["MemoryLimit"] = d.Get("memory").(int)
	payload["StorageClass"] = storageClassNvme
	payload["SpecificationClass"] = specificationClassNvme
	payload["InstanceMode"] = "Normal"
	payload["ChargeType"] = d.Get("charge_type").(string)
	payload["Quantity"] = d.Get("duration").(int)

	if v, ok := d.GetOk("vpc_id"); ok {
		payload["VPCId"] = v.(string)
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		payload["SubnetId"] = v.(string)
	}
	if v, ok := d.GetOk("tag"); ok {
		payload["Tag"] = v.(string)
	}

	resp, err := invoke(client, "CreateUDBSQLServerInstance", payload)
	if err != nil {
		return fmt.Errorf("error on creating sqlserver instance, %s", err)
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
		Target:     []string{sqlserverStateRunning},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
		Refresh:    sqlserverInstanceStateRefreshFunc(client, dbId, zone, []string{sqlserverStateRunning}),
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for sqlserver instance %q complete creating, %s", dbId, err)
	}

	return resourceUCloudSQLServerInstanceRead(d, meta)
}

func resourceUCloudSQLServerInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading sqlserver instance, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	db, err := describeDBInstanceByID(client, d.Id(), zone)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading sqlserver instance %q, %s", d.Id(), err)
	}

	d.Set("name", payloadString(db, "Name"))
	d.Set("db_version", payloadString(db, "DBTypeId"))
	d.Set("port", payloadInt(db, "Port"))
	d.Set("private_ip", payloadString(db, "VirtualIP"))
	d.Set("status", payloadString(db, "State"))
	d.Set("cpu", payloadInt(db, "CPU"))
	d.Set("memory", payloadInt(db, "MemoryLimit"))
	d.Set("disk_space", payloadInt(db, "DiskSpace"))
	d.Set("charge_type", payloadString(db, "ChargeType"))
	d.Set("vpc_id", payloadString(db, "VPCId"))
	d.Set("subnet_id", payloadString(db, "SubnetId"))
	d.Set("tag", payloadString(db, "Tag"))
	d.Set("create_time", timestampToString(payloadInt(db, "CreateTime")))
	d.Set("expire_time", timestampToString(payloadInt(db, "ExpiredTime")))
	d.Set("modify_time", timestampToString(payloadInt(db, "ModifyTime")))

	return nil
}

func resourceUCloudSQLServerInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting sqlserver instance, %s", err)
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
		if !isStringIn(state, []string{sqlserverStateShutoff, sqlserverStateFail, sqlserverStateDeleteFail, sqlserverStateRecoverFail}) {
			stopPayload := basePayload(client, zone)
			stopPayload["DBId"] = dbId
			if _, err := invoke(client, "StopUDBInstance", stopPayload); err != nil {
				return resource.RetryableError(fmt.Errorf("error on stopping sqlserver instance when deleting %q, %s", dbId, err))
			}

			stateConf := resource.StateChangeConf{
				Pending:    []string{statusPending},
				Target:     []string{sqlserverStateShutoff},
				Timeout:    d.Timeout(schema.TimeoutDelete),
				Delay:      3 * time.Second,
				MinTimeout: 2 * time.Second,
				Refresh:    sqlserverInstanceStateRefreshFunc(client, dbId, zone, []string{sqlserverStateShutoff}),
			}
			if _, err := stateConf.WaitForState(); err != nil {
				return resource.RetryableError(fmt.Errorf("error on waiting for stopping sqlserver instance when deleting %q, %s", dbId, err))
			}
		}

		deletePayload := basePayload(client, zone)
		deletePayload["DBId"] = dbId
		if _, err := invoke(client, "DeleteUDBInstance", deletePayload); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting sqlserver instance %q, %s", dbId, err))
		}

		if _, err := describeDBInstanceByID(client, dbId, zone); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading sqlserver instance when deleting %q, %s", dbId, err))
		}

		return resource.RetryableError(fmt.Errorf("the specified sqlserver instance %q has not been deleted due to unknown error", dbId))
	})
}
