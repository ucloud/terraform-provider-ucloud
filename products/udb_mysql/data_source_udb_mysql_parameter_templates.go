package udb_mysql

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceUCloudMySQLParameterTemplates() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudMySQLParameterTemplatesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"db_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(dbVersionList, false),
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

			"udb_mysql_parameter_templates": {
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
					},
				},
			},
		},
	}
}

func dataSourceUCloudMySQLParameterTemplatesRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading mysql parameter templates, %s", err)
	}
	zone := d.Get("availability_zone").(string)

	payload := basePayload(client, zone)
	if v, ok := d.GetOk("db_version"); ok {
		payload["DBVersion"] = v.(string)
	}

	resp, err := invoke(client, "ListUDBParamTemplate", payload)
	if err != nil {
		return fmt.Errorf("error on reading mysql parameter templates, %s", err)
	}

	dataSet, _ := resp["DataSet"].([]interface{})
	var templates []map[string]interface{}
	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, item := range dataSet {
			if m, ok := item.(map[string]interface{}); ok && r.MatchString(payloadString(m, "Name")) {
				templates = append(templates, m)
			}
		}
	} else {
		for _, item := range dataSet {
			if m, ok := item.(map[string]interface{}); ok {
				templates = append(templates, m)
			}
		}
	}

	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range templates {
		id := fmt.Sprintf("%d", payloadInt(item, "Id"))
		ids = append(ids, id)
		data = append(data, map[string]interface{}{
			"id":   id,
			"name": payloadString(item, "Name"),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("udb_mysql_parameter_templates", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
