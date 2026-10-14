package udb_mysql

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceUCloudMySQLInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudMySQLInstancesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"udb_mysql_instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"private_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"db_version": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"instance_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"disk_space": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"cpu": {
							Type:     schema.TypeInt,
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

						"charge_type": {
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
					},
				},
			},
		},
	}
}

func dataSourceUCloudMySQLInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading mysql instance list, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	filters := map[string]interface{}{}
	if v, ok := d.GetOk("vpc_id"); ok {
		filters["VPCId"] = v.(string)
	}

	all, err := describeDBInstances(client, zone, filters)
	if err != nil {
		return fmt.Errorf("error on reading mysql instance list, %s", err)
	}

	ids, idsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")

	var instances []map[string]interface{}
	if idsOk || nameRegexOk {
		var r *regexp.Regexp
		if nameRegexOk && nameRegex.(string) != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		idsSlice := schemaSetToStringSlice(ids)
		for _, item := range all {
			if r != nil && !r.MatchString(payloadString(item, "Name")) {
				continue
			}
			if idsOk && !isStringIn(payloadString(item, "DBId"), idsSlice) {
				continue
			}
			instances = append(instances, item)
		}
	} else {
		instances = all
	}

	if err := dataSourceUCloudMySQLInstancesSave(d, instances); err != nil {
		return fmt.Errorf("error on reading mysql instance list, %s", err)
	}
	return nil
}

func dataSourceUCloudMySQLInstancesSave(d *schema.ResourceData, instances []map[string]interface{}) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range instances {
		ids = append(ids, payloadString(item, "DBId"))
		data = append(data, map[string]interface{}{
			"id":                payloadString(item, "DBId"),
			"name":              payloadString(item, "Name"),
			"availability_zone": payloadString(item, "Zone"),
			"status":            payloadString(item, "State"),
			"private_ip":        payloadString(item, "VirtualIP"),
			"port":              payloadInt(item, "Port"),
			"db_version":        payloadString(item, "DBTypeId"),
			"instance_mode":     normalizeInstanceMode(payloadString(item, "InstanceMode")),
			"memory":            payloadInt(item, "MemoryLimit") / 1000,
			"disk_space":        payloadInt(item, "DiskSpace"),
			"cpu":               payloadInt(item, "CPU"),
			"vpc_id":            payloadString(item, "VPCId"),
			"subnet_id":         payloadString(item, "SubnetId"),
			"charge_type":       payloadString(item, "ChargeType"),
			"create_time":       timestampToString(payloadInt(item, "CreateTime")),
			"expire_time":       timestampToString(payloadInt(item, "ExpiredTime")),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("udb_mysql_instances", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
