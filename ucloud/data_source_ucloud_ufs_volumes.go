package ucloud

import (
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/ufs"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudUFSVolumes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudUFSVolumesRead,

		Schema: map[string]*schema.Schema{
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

			"ufs_volumes": {
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

						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"storage_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"protocol_type": {
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

func dataSourceUCloudUFSVolumesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ufsconn
	var allUFSs []ufs.UFSVolumeInfo2
	var ufss []ufs.UFSVolumeInfo2
	var limit int = 100
	var offset int

	for {
		req := conn.NewDescribeUFSVolume2Request()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeUFSVolume2(req)
		if err != nil {
			return fmt.Errorf("error on reading ufs list, %s", err)
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		allUFSs = append(allUFSs, resp.DataSet...)

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
		for _, v := range allUFSs {
			if r != nil && !r.MatchString(v.VolumeName) {
				continue
			}

			if idsOk && !isStringIn(v.VolumeId, schemaSetToStringSlice(ids)) {
				continue
			}
			ufss = append(ufss, v)
		}
	} else {
		ufss = allUFSs
	}

	err := dataSourceUCloudUFSVolumesSave(d, ufss)
	if err != nil {
		return fmt.Errorf("error on reading ufs list, %s", err)
	}

	return nil
}

func dataSourceUCloudUFSVolumesSave(d *schema.ResourceData, ufss []ufs.UFSVolumeInfo2) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range ufss {
		ids = append(ids, item.VolumeId)

		data = append(data, map[string]interface{}{
			"id":            item.VolumeId,
			"size":          item.Size,
			"storage_type":  item.StorageType,
			"protocol_type": item.ProtocolType,
			"name":          item.VolumeName,
			"tag":           item.Tag,
			"remark":        item.Remark,
			"create_time":   timestampToString(item.CreateTime),
			"expire_time":   timestampToString(item.ExpiredTime),
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("ufs_volumes", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
