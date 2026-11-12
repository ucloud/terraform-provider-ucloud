package udb_mysql

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceUCloudMySQLMachineTypes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudMySQLMachineTypesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
			},

			"instance_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Normal", "HA"}, false),
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"udb_mysql_machine_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
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

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"group": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"specification_class": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"storage_class": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudMySQLMachineTypesRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading mysql machine types, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	payload := basePayload(client, zone)
	if v, ok := d.GetOk("instance_mode"); ok {
		payload["InstanceMode"] = v.(string)
	}

	resp, err := invoke(client, "ListUDBMachineType", payload)
	if err != nil {
		return fmt.Errorf("error on reading mysql machine types, %s", err)
	}

	dataSet, _ := resp["DataSet"].([]interface{})
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range dataSet {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id := payloadString(m, "ID")
		ids = append(ids, id)
		data = append(data, map[string]interface{}{
			"id":                  id,
			"cpu":                 payloadInt(m, "Cpu"),
			"memory":              payloadInt(m, "Memory"),
			"description":         payloadString(m, "Description"),
			"group":               payloadString(m, "Group"),
			"specification_class": payloadString(m, "SpecificationClass"),
			"storage_class":       payloadString(m, "StorageClass"),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("udb_mysql_machine_types", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
