package ucloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudDBBackups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudDBBackupsRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"db_backups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"backup_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"backup_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"backup_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"backup_end_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"backup_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"backup_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"db_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"db_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudDBBackupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udbconn

	var allDBBackups []udb.UDBBackupSet
	var finalDBBackups []udb.UDBBackupSet
	var limit = 100
	var offset int

	for {
		req := conn.NewDescribeUDBBackupRequest()
		if val, ok := d.GetOk("availability_zone"); ok {
			req.Zone = ucloud.String(val.(string))
		}
		if val, ok := d.GetOk("project_id"); ok {
			req.ProjectId = ucloud.String(val.(string))
		}
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeUDBBackup(req)
		if err != nil {
			return fmt.Errorf("error on reading db backup list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allDBBackups = append(allDBBackups, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	nameRegex, nameRegexOk := d.GetOk("name_regex")
	if nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, v := range allDBBackups {
			if r != nil && !r.MatchString(v.BackupName) {
				continue
			}

			finalDBBackups = append(finalDBBackups, v)
		}
	} else {
		finalDBBackups = allDBBackups
	}

	err := dataSourceUCloudDBBackupsSave(d, finalDBBackups)
	if err != nil {
		return fmt.Errorf("error on reading db instance list, %s", err)
	}

	return nil
}

func dataSourceUCloudDBBackupsSave(d *schema.ResourceData, dbBackups []udb.UDBBackupSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, dbBackup := range dbBackups {

		tmBackupTime := time.Unix(int64(dbBackup.BackupTime), 0)
		tmBackupEndTime := time.Unix(int64(dbBackup.BackupEndTime), 0)
		ids = append(ids, fmt.Sprint(dbBackup.BackupId))
		data = append(data, map[string]interface{}{
			"backup_id":       dbBackup.BackupId,
			"backup_name":     dbBackup.BackupName,
			"backup_size":     dbBackup.BackupSize,
			"backup_time":     tmBackupTime.Format("2006-01-02 03:04:05"),
			"backup_end_time": tmBackupEndTime.Format("2006-01-02 03:04:05"),
			"backup_type":     dbBackup.BackupType,
			"backup_zone":     dbBackup.BackupZone,
			"db_id":           dbBackup.DBId,
			"db_name":         dbBackup.DBName,
			"state":           dbBackup.State,
			"zone":            dbBackup.Zone,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("db_backups", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
