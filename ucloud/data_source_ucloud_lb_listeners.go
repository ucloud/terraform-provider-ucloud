package ucloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudLBListeners() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLBListenersRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"lb_listeners": {
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

						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"listen_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"idle_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"method": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"persistence_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"persistence": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"health_check_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
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

func dataSourceUCloudLBListenersRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	var lbListeners []ulb.ULBVServerSet
	var limit int = 100
	var totalCount int
	var offset int

	lbId := d.Get("load_balancer_id").(string)
	if ids, ok := d.GetOk("ids"); ok {
		idSet := schemaSetToStringSlice(ids)

		for _, v := range idSet {
			vserverSet, err := client.describeVServerById(lbId, v)
			if err != nil {
				return fmt.Errorf("error on reading lb listener list, %s", err)
			}

			lbListeners = append(lbListeners, *vserverSet)
			totalCount++
		}
	} else {
		req := conn.NewDescribeVServerRequest()
		req.ULBId = ucloud.String(lbId)
		for {
			req.Limit = ucloud.Int(limit)
			req.Offset = ucloud.Int(offset)
			resp, err := conn.DescribeVServer(req)
			if err != nil {
				return fmt.Errorf("error on reading lb listener list, %s", err)
			}

			if resp == nil || len(resp.DataSet) < 1 {
				break
			}

			lbListeners = append(lbListeners, resp.DataSet...)
			totalCount = totalCount + resp.TotalCount

			if len(resp.DataSet) < limit {
				break
			}

			offset = offset + limit
		}
	}

	d.Set("total_count", totalCount)
	err := dataSourceUCloudLBListenersSave(d, lbListeners)
	if err != nil {
		return fmt.Errorf("error on reading lb listener list, %s", err)
	}

	return nil
}

func dataSourceUCloudLBListenersSave(d *schema.ResourceData, lbListeners []ulb.ULBVServerSet) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range lbListeners {
		ids = append(ids, string(item.VServerId))
		if item.MonitorType == lbPath {
			data = append(data, map[string]interface{}{
				"id":                item.VServerId,
				"name":              item.VServerName,
				"protocol":          upperCvt.convert(item.Protocol),
				"listen_type":       upperCamelCvt.convert(item.ListenType),
				"port":              item.FrontendPort,
				"idle_timeout":      item.ClientTimeout,
				"method":            upperCamelCvt.convert(item.Method),
				"persistence_type":  upperCamelCvt.convert(item.PersistenceType),
				"persistence":       item.PersistenceInfo,
				"health_check_type": upperCamelCvt.convert(item.MonitorType),
				"status":            listenerStatusCvt.convert(item.Status),
				"domain":            item.Domain,
				"path":              item.Path,
			})
		} else {
			data = append(data, map[string]interface{}{
				"id":                item.VServerId,
				"name":              item.VServerName,
				"protocol":          upperCvt.convert(item.Protocol),
				"listen_type":       upperCamelCvt.convert(item.ListenType),
				"port":              item.FrontendPort,
				"idle_timeout":      item.ClientTimeout,
				"method":            upperCamelCvt.convert(item.Method),
				"persistence_type":  upperCamelCvt.convert(item.PersistenceType),
				"persistence":       item.PersistenceInfo,
				"health_check_type": upperCamelCvt.convert(item.MonitorType),
				"status":            listenerStatusCvt.convert(item.Status),
			})
		}
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("lb_listeners", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
