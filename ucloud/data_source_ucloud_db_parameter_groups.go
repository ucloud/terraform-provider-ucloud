package ucloud

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudDBParameterGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudDBParameterGroupsRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"multi_az"},
			},

			"multi_az": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"availability_zone"},
			},

			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},

			"class_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"sql", "postgresql"}, false),
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"parameter_groups": {
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

						"name": {
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

						"is_default": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudDBParameterGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.udbconn

	req := conn.NewDescribeUDBParamGroupRequest()
	if val, ok := d.GetOkExists("class_type"); ok {
		req.ClassType = ucloud.String(val.(string))
	}

	if val, ok := d.GetOk("availability_zone"); ok {
		req.Zone = ucloud.String(val.(string))
	}
	if val, ok := d.GetOkExists("multi_az"); ok {
		req.RegionFlag = ucloud.Bool(val.(bool))
	}

	var paramGroups []udb.UDBParamGroupSet
	var allParamGroups []udb.UDBParamGroupSet
	var limit = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeUDBParamGroup(req)
		if err != nil {
			return fmt.Errorf("error on reading db parameter groups, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allParamGroups = append(allParamGroups, resp.DataSet...)

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if nameRegex, ok := d.GetOk("name_regex"); ok {
		r := regexp.MustCompile(nameRegex.(string))
		for _, v := range allParamGroups {
			if r != nil && !r.MatchString(v.GroupName) {
				continue
			}

			paramGroups = append(paramGroups, v)
		}
	} else {
		paramGroups = allParamGroups
	}

	err := dataSourceUCloudDBParameterGroupsSave(d, paramGroups)
	if err != nil {
		return fmt.Errorf("error on reading parameter group list, %s", err)
	}

	return nil
}

func dataSourceUCloudDBParameterGroupsSave(d *schema.ResourceData, parameterGroups []udb.UDBParamGroupSet) error {
	var ids []string
	var id string
	var data []map[string]interface{}
	for _, parameterGroup := range parameterGroups {
		if parameterGroup.Zone != "" {
			id = fmt.Sprintf("%s:%s", parameterGroup.Zone, strconv.Itoa(parameterGroup.GroupId))
		} else {
			id = strconv.Itoa(parameterGroup.GroupId)
		}
		ids = append(ids, id)
		arr := strings.Split(parameterGroup.DBTypeId, "-")
		data = append(data, map[string]interface{}{
			"id":                strconv.Itoa(parameterGroup.GroupId),
			"name":              parameterGroup.GroupName,
			"engine":            arr[0],
			"engine_version":    arr[1],
			"is_default":        !parameterGroup.Modifiable,
			"availability_zone": parameterGroup.Zone,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	if err := d.Set("parameter_groups", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
