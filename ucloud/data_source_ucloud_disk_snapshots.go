package ucloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
)

func dataSourceUCloudDiskSnapshots() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudDiskSnapshotsRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"disk_id": {
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

			"snapshots": {
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

						"disk_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"disk_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"source_disk_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"is_disk_available": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudDiskSnapshotsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	zone := ""
	if v, ok := d.GetOk("availability_zone"); ok {
		zone = v.(string)
	}

	diskId := ""
	if v, ok := d.GetOk("disk_id"); ok {
		diskId = v.(string)
	}

	// the remote api requires the zone to be specified when filtering snapshots by disk
	if diskId != "" && zone == "" {
		return fmt.Errorf("error on reading disk snapshot list, the %q must be specified when %q is set", "availability_zone", "disk_id")
	}

	allSnapshots, err := client.describeSnapshots(zone, diskId)
	if err != nil {
		return fmt.Errorf("error on reading disk snapshot list, %s", err)
	}

	var snapshots []udisk.UDiskSnapshotSet
	ids, idsOk := d.GetOk("ids")
	nameRegex, nameRegexOk := d.GetOk("name_regex")
	if idsOk || nameRegexOk {
		var r *regexp.Regexp
		if nameRegex != "" {
			r = regexp.MustCompile(nameRegex.(string))
		}
		for _, v := range allSnapshots {
			if r != nil && !r.MatchString(v.Name) {
				continue
			}

			if idsOk && !isStringIn(v.SnapshotId, schemaSetToStringSlice(ids)) {
				continue
			}
			snapshots = append(snapshots, v)
		}
	} else {
		snapshots = allSnapshots
	}

	err = dataSourceUCloudDiskSnapshotsSave(d, snapshots)
	if err != nil {
		return fmt.Errorf("error on reading disk snapshot list, %s", err)
	}

	return nil
}

func dataSourceUCloudDiskSnapshotsSave(d *schema.ResourceData, snapshots []udisk.UDiskSnapshotSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range snapshots {
		ids = append(ids, item.SnapshotId)

		data = append(data, map[string]interface{}{
			"id":                item.SnapshotId,
			"availability_zone": item.Zone,
			"disk_id":           item.UDiskId,
			"name":              item.Name,
			"comment":           item.Comment,
			"disk_type":         snapshotDiskTypeCvt.convert(item.DiskType),
			"size":              item.Size,
			"source_disk_name":  item.UDiskName,
			"is_disk_available": item.IsUDiskAvailable,
			"create_time":       timestampToString(item.CreateTime),
			"status":            item.Status,
		})
	}

	d.SetId(hashStringArray(ids))
	d.Set("total_count", len(data))
	d.Set("ids", ids)
	if err := d.Set("snapshots", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
