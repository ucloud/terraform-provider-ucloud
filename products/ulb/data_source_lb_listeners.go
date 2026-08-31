package ulb

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"load_balancer_id": {Type: schema.TypeString, Required: true},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"output_file": {Type: schema.TypeString, Optional: true},
			"total_count": {Type: schema.TypeInt, Computed: true},
			"lb_listeners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":                {Type: schema.TypeString, Computed: true},
					"name":              {Type: schema.TypeString, Computed: true},
					"protocol":          {Type: schema.TypeString, Computed: true},
					"listen_type":       {Type: schema.TypeString, Computed: true},
					"port":              {Type: schema.TypeInt, Computed: true},
					"idle_timeout":      {Type: schema.TypeInt, Computed: true},
					"method":            {Type: schema.TypeString, Computed: true},
					"persistence_type":  {Type: schema.TypeString, Computed: true},
					"persistence":       {Type: schema.TypeString, Computed: true},
					"health_check_type": {Type: schema.TypeString, Computed: true},
					"domain":            {Type: schema.TypeString, Computed: true},
					"path":              {Type: schema.TypeString, Computed: true},
					"status":            {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudLBListenersRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb listener list, %s", err)
	}
	conn := client.ulbconn
	var allListeners []ulb.ULBVServerSet
	const limit = 100
	lbID := d.Get("load_balancer_id").(string)
	for offset := 0; ; offset += limit {
		req := conn.NewDescribeVServerRequest()
		req.ULBId = ucloud.String(lbID)
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeVServer(req)
		if err != nil {
			return fmt.Errorf("error on reading lb listener list, %s", err)
		}
		if resp == nil || len(resp.DataSet) < 1 {
			break
		}
		allListeners = append(allListeners, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
	}

	ids, idsOK := d.GetOk("ids")
	nameRegex, regexOK := d.GetOk("name_regex")
	listeners := allListeners
	if idsOK || regexOK {
		listeners = nil
		var re *regexp.Regexp
		if regexOK {
			re = regexp.MustCompile(nameRegex.(string))
		}
		for _, item := range allListeners {
			if re != nil && !re.MatchString(item.VServerName) {
				continue
			}
			if idsOK && !isStringIn(item.VServerId, schemaSetToStringSlice(ids)) {
				continue
			}
			listeners = append(listeners, item)
		}
	}
	if err := dataSourceUCloudLBListenersSave(d, listeners); err != nil {
		return fmt.Errorf("error on reading lb listener list, %s", err)
	}
	return nil
}

func dataSourceUCloudLBListenersSave(d *schema.ResourceData, listeners []ulb.ULBVServerSet) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range listeners {
		ids = append(ids, item.VServerId)
		value := map[string]interface{}{
			"id": item.VServerId, "name": item.VServerName,
			"protocol": upperCvt.convert(item.Protocol), "listen_type": upperCamelConvert(item.ListenType),
			"port": item.FrontendPort, "idle_timeout": item.ClientTimeout,
			"method": upperCamelConvert(item.Method), "persistence_type": upperCamelConvert(item.PersistenceType),
			"persistence": item.PersistenceInfo, "health_check_type": upperCamelConvert(item.MonitorType),
			"status": listenerStatusCvt.convert(item.Status),
		}
		if item.MonitorType == lbMatchTypePath {
			value["domain"] = item.Domain
			value["path"] = item.Path
		}
		data = append(data, value)
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("lb_listeners", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
