package umongodb

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceUCloudUMongoDBBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUMongoDBBackupCreate,
		Read:   resourceUCloudUMongoDBBackupRead,
		Delete: resourceUCloudUMongoDBBackupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUCloudUMongoDBBackupImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},
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
			"backup_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"backup_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"end_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceUCloudUMongoDBBackupCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating umongodb backup, %s", err)
	}
	clusterID := data.Get("cluster_id").(string)
	zone := data.Get("availability_zone").(string)
	backupName := data.Get("backup_name").(string)
	createdAfter := int(time.Now().Add(-2 * time.Minute).Unix())
	backupID, err := createBackup(client, clusterID, zone, backupName)
	if err != nil {
		return err
	}
	data.SetId(backupID)

	confTimeout := data.Timeout(schema.TimeoutCreate)
	deadline := time.Now().Add(confTimeout)
	for {
		info, readErr := findCreatedBackup(client, clusterID, zone, data.Id(), backupName, createdAfter)
		if readErr != nil && !isNotFoundError(readErr) {
			return fmt.Errorf("error on reading umongodb backup %q, %s", data.Id(), readErr)
		}
		if info != nil {
			if info.BackupId != "" && info.BackupId != data.Id() {
				data.SetId(info.BackupId)
			}
			if backupCreateFinished(info.State) {
				if info.State != "Backup_Success" {
					return fmt.Errorf("umongodb backup %s entered %s", data.Id(), info.State)
				}
				return resourceUCloudUMongoDBBackupRead(data, meta)
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("wait for umongodb backup %s timed out; a sharded failed backup is omitted from the backup list", data.Id())
		}
		time.Sleep(15 * time.Second)
	}
}

func resourceUCloudUMongoDBBackupRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading umongodb backup %q, %s", data.Id(), err)
	}
	info, err := findBackup(client, data.Get("cluster_id").(string), data.Get("availability_zone").(string), data.Id(), data.Get("backup_name").(string), data.Get("start_time").(int))
	if err != nil {
		if isNotFoundError(err) {
			data.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading umongodb backup %q, %s", data.Id(), err)
	}
	if info.BackupId != "" && info.BackupId != data.Id() {
		data.SetId(info.BackupId)
	}
	_ = data.Set("cluster_id", info.ClusterId)
	_ = data.Set("backup_name", info.BackupName)
	_ = data.Set("state", info.State)
	_ = data.Set("backup_size", info.BackupSize)
	_ = data.Set("backup_type", info.BackupType)
	_ = data.Set("start_time", info.StartTime)
	_ = data.Set("end_time", info.EndTime)
	return nil
}

func resourceUCloudUMongoDBBackupDelete(data *schema.ResourceData, meta interface{}) error {
	data.SetId("")
	return nil
}

func resourceUCloudUMongoDBBackupImport(data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(data.Id(), "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%s), expected <availability_zone>/<cluster_id>/<backup_id>", data.Id())
	}
	if err := data.Set("availability_zone", parts[0]); err != nil {
		return nil, err
	}
	if err := data.Set("cluster_id", parts[1]); err != nil {
		return nil, err
	}
	data.SetId(parts[2])
	return []*schema.ResourceData{data}, nil
}

func backupCreateFinished(state string) bool {
	switch state {
	case "Backup_Success", "Backup_Failed", "Backup_Stopped", "Backup_StopFailed":
		return true
	default:
		return false
	}
}
