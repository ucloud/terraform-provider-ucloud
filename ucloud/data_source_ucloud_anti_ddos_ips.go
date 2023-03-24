package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/uads"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudAntiDDoSIPs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudAntiDDoSIPsRead,

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"proxy_ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"total_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceUCloudAntiDDoSIPsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).uadsconn
	var allIPs []uads.GameIpInfoTotal
	var limit int = 100
	var offset int
	instanceId := d.Get("instance_id").(string)

	for {
		req := conn.NewDescribeHighProtectGameIPInfoRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		req.ResourceId = ucloud.String(instanceId)
		resp, err := conn.DescribeHighProtectGameIPInfo(req)
		if err != nil {
			return fmt.Errorf("error on reading anti ddos ip list, %s", err)
		}

		if resp == nil || len(resp.GameIPInfo) < 1 {
			break
		}

		allIPs = append(allIPs, resp.GameIPInfo...)

		if len(resp.GameIPInfo) < limit {
			break
		}

		offset = offset + limit
	}
	var ips []interface{}
	for _, ipInfo := range allIPs {
		ip := make(map[string]interface{})
		ip["instance_id"] = instanceId
		ip["ip"] = ipInfo.DefenceIP
		ip["comment"] = ipInfo.Remark
		ip["domain"] = ipInfo.Cname
		ip["status"] = ipInfo.Status
		ip["proxy_ips"] = ipInfo.SrcIP
		ips = append(ips, ip)
	}
	d.SetId(instanceId)
	d.Set("total_count", len(ips))
	d.Set("ips", ips)
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), ips)
	}
	return nil
}
