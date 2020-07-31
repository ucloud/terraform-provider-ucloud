package ucloud

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudDBInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudDBInstancesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
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

			"db_instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"standby_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"engine": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"private_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"instance_storage": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"charge_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"backup_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"backup_date": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"backup_begin_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"backup_black_list": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},

						"tag": {
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
				},
			},
		},
	}
}

func dataSourceUCloudDBInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udbconn

	var allDBInstances []udb.UDBInstanceSet
	var dbInstances []udb.UDBInstanceSet
	var limit int = 100
	var offset int

	for {
		req := conn.NewDescribeUDBInstanceRequest()
		if val, ok := d.GetOk("availability_zone"); ok {
			req.Zone = ucloud.String(val.(string))
		}
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		req.ClassType = ucloud.String("SQL")

		resp, err := conn.DescribeUDBInstance(req)
		if err != nil {
			return fmt.Errorf("error on reading db instance list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allDBInstances = append(allDBInstances, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	ids, idsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")
	if idsOk || nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, v := range allDBInstances {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			if idsOk && !isStringIn(v.DBId, schemaSetToStringSlice(ids)) {
				continue
			}
			dbInstances = append(dbInstances, v)
		}
	} else {
		dbInstances = allDBInstances
	}

	err := dataSourceUCloudDBInstancesSave(d, dbInstances)
	if err != nil {
		return fmt.Errorf("error on reading db instance list, %s", err)
	}

	return nil
}

func dataSourceUCloudDBInstancesSave(d *schema.ResourceData, dbInstances []udb.UDBInstanceSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, dbInstance := range dbInstances {

		ids = append(ids, dbInstance.DBId)
		backupBlackList := strings.Split(dbInstance.BackupBlacklist, ";")
		arr := strings.Split(dbInstance.DBTypeId, "-")
		dbType := dbInstanceType{}
		dbType.Memory = dbInstance.MemoryLimit / 1000
		dbType.Engine = arr[0]
		dbType.Mode = dbModeCvt.unconvert(dbInstance.InstanceMode)
		dbType.Type = dbTypeCvt.unconvert(dbInstance.InstanceType)
		instanceType := fmt.Sprintf("%s-%s-%d", dbType.Engine, dbType.Mode, dbType.Memory)
		if dbType.Type == dbNVMeInstanceType {
			instanceType = fmt.Sprintf("%s-%s-%s-%d", dbType.Engine, dbType.Mode, dbType.Type, dbType.Memory)
		}

		data = append(data, map[string]interface{}{
			"id":                dbInstance.DBId,
			"availability_zone": dbInstance.Zone,
			"instance_type":     instanceType,
			"standby_zone":      dbInstance.BackupZone,
			"name":              dbInstance.Name,
			"vpc_id":            dbInstance.VPCId,
			"subnet_id":         dbInstance.SubnetId,
			"engine":            arr[0],
			"engine_version":    arr[1],
			"port":              dbInstance.Port,
			"private_ip":        dbInstance.VirtualIP,
			"status":            dbInstance.State,
			"instance_storage":  dbInstance.DiskSpace,
			"charge_type":       upperCamelCvt.convert(dbInstance.ChargeType),
			"backup_count":      dbInstance.BackupCount,
			"backup_date":       dbInstance.BackupDate,
			"backup_begin_time": dbInstance.BackupBeginTime,
			"backup_black_list": backupBlackList,
			"tag":               dbInstance.Tag,
			"create_time":       timestampToString(dbInstance.CreateTime),
			"expire_time":       timestampToString(dbInstance.ExpiredTime),
			"modify_time":       timestampToString(dbInstance.ModifyTime),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("db_instances", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
