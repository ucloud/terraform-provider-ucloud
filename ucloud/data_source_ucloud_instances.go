package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudInstancesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"ids": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},

			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ForceNew: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},

			"instances": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"cpu": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"memory": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},

						"create_time": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"expire_time": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"instance_charge_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"auto_renew": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"remark": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"tag": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},

						"disk_set": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_type": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},

									"size": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},

									"disk_id": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},

									"is_boot": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"ip_set": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},

									"type": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudInstancesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uhostconn

	req := conn.NewDescribeUHostInstanceRequest()

	if ids, ok := d.GetOk("ids"); ok && len(ids.([]interface{})) > 0 {
		req.UHostIds = ifaceToStringSlice(ids)
	}

	var fetched []uhost.UHostInstanceSet
	var limit int = 100
	var offset int
	for {
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeUHostInstance(req)
		if err != nil {
			return fmt.Errorf("error in read instance list, %s", err)
		}

		if resp == nil || len(resp.UHostSet) < 1 {
			break
		}

		fetched = append(fetched, resp.UHostSet...)

		if len(resp.UHostSet) < limit {
			break
		}

		offset = offset + limit
	}

	tags, tagOk := d.GetOk("tags")
	var instances []uhost.UHostInstanceSet
	var totalCount int
	for _, item := range fetched {

		if tagOk && checkStringIn(item.Tag, tags.([]string)) != nil {
			continue
		}

		instances = append(instances, item)
		totalCount = totalCount + 1
	}

	d.Set("total_count", totalCount)
	err := dataSourceUCloudInstancesSave(d, instances)
	if err != nil {
		return fmt.Errorf("error in read instance list, %s", err)
	}

	return nil
}

func dataSourceUCloudInstancesSave(d *schema.ResourceData, instances []uhost.UHostInstanceSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, instance := range instances {
		ids = append(ids, string(instance.UHostId))
		ipSet := []map[string]interface{}{}
		for _, item := range instance.IPSet {
			ipSet = append(ipSet, map[string]interface{}{
				"ip":   item.IP,
				"type": item.Type,
			})
		}

		diskSet := []map[string]interface{}{}
		for _, item := range instance.DiskSet {
			diskSet = append(diskSet, map[string]interface{}{
				"disk_type": item.DiskType,
				"size":      item.Size,
				"disk_id":   item.DiskId,
				"is_boot":   item.IsBoot,
			})
		}

		data = append(data, map[string]interface{}{
			"availability_zone":    instance.Zone,
			"id":                   instance.UHostId,
			"name":                 instance.Name,
			"cpu":                  instance.CPU,
			"memory":               instance.Memory,
			"create_time":          timestampToString(instance.CreateTime),
			"expire_time":          timestampToString(instance.ExpireTime),
			"auto_renew":           instance.AutoRenew,
			"remark":               instance.Remark,
			"tag":                  instance.Tag,
			"status":               instance.State,
			"instance_charge_type": instance.ChargeType,
			"ip_set":               ipSet,
			"disk_set":             diskSet,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("instances", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
