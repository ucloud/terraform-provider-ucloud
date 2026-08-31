package ulb

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func dataSourceUCloudLBSSLs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLBSSLsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"output_file": {Type: schema.TypeString, Optional: true},
			"total_count": {Type: schema.TypeInt, Computed: true},
			"lb_ssls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":          {Type: schema.TypeString, Computed: true},
					"name":        {Type: schema.TypeString, Computed: true},
					"create_time": {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudLBSSLsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb ssl list, %s", err)
	}
	conn := client.ulbconn
	var allSSLs []ulb.ULBSSLSet
	const limit = 100
	for offset := 0; ; offset += limit {
		req := conn.NewDescribeSSLRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)
		resp, err := conn.DescribeSSL(req)
		if err != nil {
			return fmt.Errorf("error on reading lb ssl list, %s", err)
		}
		if resp == nil || len(resp.DataSet) < 1 {
			break
		}
		allSSLs = append(allSSLs, resp.DataSet...)
		if len(resp.DataSet) < limit {
			break
		}
	}
	ids, idsOK := d.GetOk("ids")
	nameRegex, regexOK := d.GetOk("name_regex")
	var ssls []ulb.ULBSSLSet
	if idsOK || regexOK {
		var re *regexp.Regexp
		if regexOK {
			re = regexp.MustCompile(nameRegex.(string))
		}
		for _, item := range allSSLs {
			if re != nil && !re.MatchString(item.SSLName) {
				continue
			}
			if idsOK && !isStringIn(item.SSLId, schemaSetToStringSlice(ids)) {
				continue
			}
			ssls = append(ssls, item)
		}
	} else {
		ssls = allSSLs
	}
	if err := dataSourceUCloudLBSSLsSave(d, ssls); err != nil {
		return fmt.Errorf("error on reading lb ssl list, %s", err)
	}
	return nil
}

func dataSourceUCloudLBSSLsSave(d *schema.ResourceData, ssls []ulb.ULBSSLSet) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range ssls {
		ids = append(ids, item.SSLId)
		data = append(data, map[string]interface{}{
			"id": item.SSLId, "name": item.SSLName, "create_time": timestampToString(item.CreateTime),
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("lb_ssls", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
