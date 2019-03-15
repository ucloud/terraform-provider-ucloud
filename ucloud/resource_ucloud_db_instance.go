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
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudDBInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudDBInstanceCreate,
		Read:   resourceUCloudDBInstanceRead,
		Update: resourceUCloudDBInstanceUpdate,
		Delete: resourceUCloudDBInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.All(
			diffValidateDBMemoryWithInstanceStorage,
			diffValidateDBEngineAndEngineVersion,
			diffValidateDBStandbyZone,
			customdiff.ValidateChange("instance_type", diffValidateDBInstanceType),
		),

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"standby_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ValidateFunc: validateDBInstancePassword,
			},

			"engine": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"mysql", "percona"}, false),
				ForceNew:     true,
				Required:     true,
			},

			"engine_version": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"5.5", "5.6", "5.7"}, false),
				ForceNew:     true,
				Required:     true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateDBInstanceName,
			},

			"instance_storage": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validateAll(
					validation.IntBetween(20, 3000),
					validateMod(10),
				),
			},

			"instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateDBInstanceType,
			},

			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(3306, 65535),
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "month",
				ValidateFunc: validation.StringInSlice([]string{
					"month",
					"year",
					"dynamic",
				}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"backup_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  7,
				ForceNew: true,
			},

			"backup_begin_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 23),
			},

			"backup_date": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"backup_black_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateDBInstanceBlackList,
				},
				Set: schema.HashString,
			},

			"tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateTag,
				Computed:     true,
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

func resourceUCloudDBInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udbconn

	engine := d.Get("engine").(string)
	// skip error because it has been validated by schema
	dbType, _ := parseDBInstanceType(d.Get("instance_type").(string))
	zone := d.Get("availability_zone").(string)
	engineVersion := d.Get("engine_version").(string)
	dbTypeId := strings.Join([]string{engine, engineVersion}, "-")

	req := conn.NewCreateUDBInstanceRequest()
	req.ChargeType = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_type").(string)))
	req.Zone = ucloud.String(zone)
	req.DiskSpace = ucloud.Int(d.Get("instance_storage").(int))
	req.AdminUser = ucloud.String("root")
	req.InstanceType = ucloud.String("SATA_SSD")
	req.MemoryLimit = ucloud.Int(dbType.Memory * 1000)
	req.InstanceMode = ucloud.String(dbModeCvt.convert(dbType.Type))
	req.DBTypeId = ucloud.String(dbTypeId)
	req.BackupCount = ucloud.Int(d.Get("backup_count").(int))
	password := fmt.Sprintf("%s%s%s",
		acctest.RandStringFromCharSet(5, defaultPasswordStr),
		acctest.RandStringFromCharSet(1, defaultPasswordSpe),
		acctest.RandStringFromCharSet(5, defaultPasswordNum))
	if v, ok := d.GetOk("password"); ok {
		req.AdminPassword = ucloud.String(v.(string))
	} else {
		req.AdminPassword = ucloud.String(password)
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = ucloud.String(v.(string))
	} else {
		req.Name = ucloud.String(resource.PrefixedUniqueId("tf-db-instance-"))
	}

	if v, ok := d.GetOk("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	// if tag is empty string, use default tag
	if v, ok := d.GetOk("tag"); ok {
		req.Tag = ucloud.String(v.(string))
	} else {
		req.Tag = ucloud.String(defaultTag)
	}

	if val, ok := d.GetOk("port"); ok {
		req.Port = ucloud.Int(val.(int))
	} else {
		if engine == "mysql" || engine == "percona" {
			req.Port = ucloud.Int(3306)
		}
	}

	if val, ok := d.GetOk("standby_zone"); ok {
		req.BackupZone = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(val.(string))
	}

	// set default value of parametergroup
	parameterGroupId, err := setDefaultParameterGroup(d, conn, zone, dbTypeId, engine, engineVersion)
	if err != nil {
		return err
	} else {
		req.ParamGroupId = ucloud.Int(parameterGroupId)
	}

	resp, err := conn.CreateUDBInstance(req)
	if err != nil {
		return fmt.Errorf("error on creating db instance, %s", err)
	}

	d.SetId(resp.DBId)
	if _, ok := d.GetOk("password"); !ok {
		d.Set("password", password)
	}

	// after create db, we need to wait it initialized
	stateConf := client.dbWaitForState(d.Id(), []string{"Running"})

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for db instance %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudDBInstanceUpdate(d, meta)
}

func resourceUCloudDBInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udbconn

	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewModifyUDBInstanceNameRequest()
		req.DBId = ucloud.String(d.Id())
		req.Name = ucloud.String(d.Get("name").(string))

		if _, err := conn.ModifyUDBInstanceName(req); err != nil {
			return fmt.Errorf("error on %s to db instance %q, %s", "ModifyUDBInstanceName", d.Id(), err)
		}
		d.SetPartial("name")
	}

	if d.HasChange("password") && !d.IsNewResource() {
		req := conn.NewModifyUDBInstancePasswordRequest()
		req.DBId = ucloud.String(d.Id())
		req.Password = ucloud.String(d.Get("password").(string))

		if _, err := conn.ModifyUDBInstancePassword(req); err != nil {
			return fmt.Errorf("error on %s to db instance %q, %s", "ModifyUDBInstancePassword", d.Id(), err)
		}
		d.SetPartial("password")
	}

	isSizeChanged := false
	sizeReq := conn.NewResizeUDBInstanceRequest()
	sizeReq.DBId = ucloud.String(d.Id())
	dbType, _ := parseDBInstanceType(d.Get("instance_type").(string))
	memory := dbType.Memory
	instanceStorage := d.Get("instance_storage").(int)

	if d.HasChange("instance_type") && !d.IsNewResource() {
		sizeReq.MemoryLimit = ucloud.Int(memory * 1000)
		isSizeChanged = true
	}

	if d.HasChange("instance_storage") && !d.IsNewResource() {
		sizeReq.DiskSpace = ucloud.Int(instanceStorage)
		sizeReq.InstanceType = ucloud.String("SATA_SSD")
		isSizeChanged = true
	}

	if isSizeChanged {
		if _, err := conn.ResizeUDBInstance(sizeReq); err != nil {
			return fmt.Errorf("error on %s to db instance %q, %s", "ResizeUDBInstance", d.Id(), err)
		}

		// after resize db instance, we need to wait it completed
		stateConf := client.dbWaitForState(d.Id(), []string{"Running", "Shutoff"})

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error on waiting for resizing db instance when updating %q, %s", d.Id(), err)
		}

		d.SetPartial("instance_storage")
		d.SetPartial("instance_type")
	}

	backupChanged := false
	buReq := conn.NewUpdateUDBInstanceBackupStrategyRequest()
	buReq.DBId = ucloud.String(d.Id())

	if d.HasChange("backup_date") {
		buReq.BackupDate = ucloud.String(d.Get("backup_date").(string))
		backupChanged = true
	}

	if d.HasChange("backup_begin_time") {
		buReq.BackupTime = ucloud.Int(d.Get("backup_begin_time").(int))
		backupChanged = true
	}

	if backupChanged {
		if _, err := conn.UpdateUDBInstanceBackupStrategy(buReq); err != nil {
			return fmt.Errorf("error on %s to db instance %q, %s", "UpdateUDBInstanceBackupStrategy", d.Id(), err)
		}

		d.SetPartial("backup_date")
		d.SetPartial("backup_begin_time")
	}

	if d.HasChange("backup_black_list") {
		blackList := strings.Join(schemaListToStringSlice(d.Get("backup_black_list").(*schema.Set).List()), ";")
		req := conn.NewEditUDBBackupBlacklistRequest()
		req.Blacklist = ucloud.String(blackList)
		req.DBId = ucloud.String(d.Id())

		if _, err := conn.EditUDBBackupBlacklist(req); err != nil {
			return fmt.Errorf("error on %s to db instance %q, %s", "EditUDBBackupBlacklist", d.Id(), err)
		}

		d.SetPartial("backup_black_list")
	}

	d.Partial(false)

	return resourceUCloudDBInstanceRead(d, meta)
}

func resourceUCloudDBInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	db, err := client.describeDBInstanceById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading db instance %q, %s", d.Id(), err)
	}

	arr := strings.Split(db.DBTypeId, "-")
	backupBlackList := strings.Split(db.BackupBlacklist, ";")
	dbType := dbInstanceType{}
	dbType.Memory = db.MemoryLimit / 1000
	dbType.Engine = arr[0]
	dbType.Type = dbModeCvt.unconvert(db.InstanceMode)

	d.Set("name", db.Name)
	d.Set("engine", arr[0])
	d.Set("engine_version", arr[1])
	d.Set("port", db.Port)
	d.Set("status", db.State)
	d.Set("password", d.Get("password"))
	d.Set("charge_type", upperCamelCvt.convert(db.ChargeType))
	d.Set("instance_storage", db.DiskSpace)
	d.Set("standby_zone", db.BackupZone)
	d.Set("availability_zone", db.Zone)
	d.Set("backup_count", db.BackupCount)
	d.Set("backup_begin_time", db.BackupBeginTime)
	d.Set("backup_date", db.BackupDate)
	d.Set("tag", db.Tag)
	d.Set("vpc_id", db.VPCId)
	d.Set("subnet_id", db.SubnetId)
	d.Set("create_time", timestampToString(db.CreateTime))
	d.Set("expire_time", timestampToString(db.ExpiredTime))
	d.Set("modify_time", timestampToString(db.ModifyTime))
	d.Set("backup_black_list", backupBlackList)
	d.Set("instance_type", fmt.Sprintf("%s-%s-%d", dbType.Engine, dbType.Type, dbType.Memory))

	return nil
}

func resourceUCloudDBInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udbconn

	req := conn.NewDeleteUDBInstanceRequest()
	req.DBId = ucloud.String(d.Id())
	stopReq := conn.NewStopUDBInstanceRequest()
	stopReq.DBId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		db, err := client.describeDBInstanceById(d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if db.State != "Shutoff" {
			if _, err := conn.StopUDBInstance(stopReq); err != nil {
				return resource.RetryableError(fmt.Errorf("error on stopping db instance when deleting %q, %s", d.Id(), err))
			}

			// after instance stop, we need to wait it stoped
			stateConf := client.dbWaitForState(d.Id(), []string{"Shutoff"})

			if _, err := stateConf.WaitForState(); err != nil {
				return resource.RetryableError(fmt.Errorf("error on waiting for stopping db instance when deleting %q, %s", d.Id(), err))
			}
		}

		if _, err := conn.DeleteUDBInstance(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting db instance %q, %s", d.Id(), err))
		}

		if _, err := client.describeDBInstanceById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading db instance when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified db instance %q has not been deleted due to unknown error", d.Id()))
	})
}

func setDefaultParameterGroup(d *schema.ResourceData, conn *udb.UDBClient, zone, dbTypeId, engine, engineVersion string) (int, error) {
	limit := 100
	offset := 0
	parameterGroupId := 0
	for {
		pgReq := conn.NewDescribeUDBParamGroupRequest()
		pgReq.Limit = ucloud.Int(limit)
		pgReq.Offset = ucloud.Int(offset)
		pgReq.Zone = ucloud.String(zone)

		if val, ok := d.GetOk("standby_zone"); ok && val.(string) != zone {
			pgReq.RegionFlag = ucloud.Bool(true)
		}

		resp, err := conn.DescribeUDBParamGroup(pgReq)
		if err != nil {
			return 0, fmt.Errorf("error on reading db parameter group when creating, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			return 0, fmt.Errorf("error on querying defult value of db parameter group")
		}

		for _, item := range resp.DataSet {
			if item.DBTypeId == dbTypeId && item.Modifiable == false {
				parameterGroupId = item.GroupId
				return parameterGroupId, nil
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}
	return 0, fmt.Errorf("can not get the default parameter group")
}

func diffValidateDBInstanceType(old, new, meta interface{}) error {
	if len(old.(string)) > 0 {
		oldType, _ := parseDBInstanceType(old.(string))
		newType, _ := parseDBInstanceType(new.(string))
		if newType.Type != oldType.Type {
			return fmt.Errorf("db instance is not supported update the type of %q", "instance_type")
		}
	}

	return nil
}

func diffValidateDBMemoryWithInstanceStorage(diff *schema.ResourceDiff, v interface{}) error {
	dbType, _ := parseDBInstanceType(diff.Get("instance_type").(string))
	memory := dbType.Memory
	instanceStorage := diff.Get("instance_storage").(int)

	if memory <= 8 && instanceStorage > 500 {
		return fmt.Errorf("the upper limit of %q is 500 when the memory is 8 or less", "instance_storage")
	}

	if memory <= 24 && instanceStorage > 1000 {
		return fmt.Errorf("the upper limit of %q is 1000 when the memory between 12 and 24", "instance_storage")
	}

	if memory == 32 && instanceStorage > 2000 {
		return fmt.Errorf("the upper limit of %q is 2000 when the memory is 32", "instance_storage")
	}

	return nil
}

func diffValidateDBEngineAndEngineVersion(diff *schema.ResourceDiff, v interface{}) error {
	engine := diff.Get("engine").(string)
	engineVersion := diff.Get("engine_version").(string)
	dbType, _ := parseDBInstanceType(diff.Get("instance_type").(string))

	if err := checkStringIn(engineVersion, []string{"5.5", "5.6", "5.7"}); err != nil && (engine == "mysql" || engine == "percona") {
		return fmt.Errorf("The current engine version %q is not supported, %s", engineVersion, err)
	}

	if dbType.Engine != engine {
		return fmt.Errorf("engine of instance type %q must be same as engine %q", dbType.Engine, engine)
	}

	return nil
}

func diffValidateDBStandbyZone(diff *schema.ResourceDiff, v interface{}) error {
	zone := diff.Get("availability_zone").(string)

	if val, ok := diff.GetOk("standby_zone"); ok && val.(string) == zone {
		return fmt.Errorf("standby_zone %q must be different from availability_zone %q", val.(string), zone)
	}

	return nil
}

func (client *UCloudClient) dbWaitForState(dbId string, target []string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     target,
		Timeout:    5 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 2 * time.Second,
		Refresh: func() (interface{}, string, error) {
			db, err := client.describeDBInstanceById(dbId)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			state := db.State
			if state == "RecoverFail" {
				return nil, "", fmt.Errorf("db instance recover failed, please make sure your %q is correct and matched with the other parameters", "backup_id")
			}

			if !isStringIn(state, target) {
				state = statusPending
			}

			return db, state, nil
		},
	}
}
