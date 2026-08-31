package ufs

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	ufsapi "github.com/ucloud/ucloud-sdk-go/services/ufs"
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

func dataSourceUCloudUFSVolumesRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading ufs list, %s", err)
	}

	const limit = 100
	var allVolumes []ufsapi.UFSVolumeInfo2
	for offset := 0; ; offset += limit {
		request := client.NewDescribeUFSVolume2Request()
		request.Limit = ucloud.Int(limit)
		request.Offset = ucloud.Int(offset)
		response, err := client.DescribeUFSVolume2(request)
		if err != nil {
			return fmt.Errorf("error on reading ufs list, %s", err)
		}
		if response == nil || len(response.DataSet) < 1 {
			break
		}
		allVolumes = append(allVolumes, response.DataSet...)
		if len(response.DataSet) < limit {
			break
		}
	}

	volumes := allVolumes
	ids, idsOK := data.GetOk("ids")
	nameRegex, nameRegexOK := data.GetOk("name_regex")
	if idsOK || nameRegexOK {
		var matcher *regexp.Regexp
		if nameRegexOK && nameRegex.(string) != "" {
			matcher = regexp.MustCompile(nameRegex.(string))
		}
		var requestedIDs []string
		if idsOK {
			requestedIDs = schemaSetToStringSlice(ids)
		}
		volumes = nil
		for _, volume := range allVolumes {
			if matcher != nil && !matcher.MatchString(volume.VolumeName) {
				continue
			}
			if idsOK && !isStringIn(volume.VolumeId, requestedIDs) {
				continue
			}
			volumes = append(volumes, volume)
		}
	}

	if err := dataSourceUCloudUFSVolumesSave(data, volumes); err != nil {
		return fmt.Errorf("error on reading ufs list, %s", err)
	}
	return nil
}

func dataSourceUCloudUFSVolumesSave(data *schema.ResourceData, volumes []ufsapi.UFSVolumeInfo2) error {
	ids := make([]string, 0, len(volumes))
	values := make([]map[string]interface{}, 0, len(volumes))
	for _, volume := range volumes {
		ids = append(ids, volume.VolumeId)
		values = append(values, map[string]interface{}{
			"id":            volume.VolumeId,
			"size":          volume.Size,
			"storage_type":  volume.StorageType,
			"protocol_type": volume.ProtocolType,
			"name":          volume.VolumeName,
			"tag":           volume.Tag,
			"remark":        volume.Remark,
			"create_time":   timestampToString(volume.CreateTime),
			"expire_time":   timestampToString(volume.ExpiredTime),
		})
	}

	data.SetId(hashStringArray(ids))
	data.Set("total_count", len(values))
	data.Set("ids", ids)
	if err := data.Set("ufs_volumes", values); err != nil {
		return err
	}
	if outputFile, ok := data.GetOk("output_file"); ok && outputFile.(string) != "" {
		if err := writeToFile(outputFile.(string), values); err != nil {
			return err
		}
	}
	return nil
}
