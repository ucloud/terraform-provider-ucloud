package ulb

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
)

func dataSourceUCloudLBAttachments() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudLBAttachmentsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"load_balancer_id": {Type: schema.TypeString, Required: true},
			"listener_id":      {Type: schema.TypeString, Required: true},
			"output_file":      {Type: schema.TypeString, Optional: true},
			"total_count":      {Type: schema.TypeInt, Computed: true},
			"lb_attachments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":          {Type: schema.TypeString, Computed: true},
					"resource_id": {Type: schema.TypeString, Computed: true},
					"port":        {Type: schema.TypeInt, Computed: true},
					"private_ip":  {Type: schema.TypeString, Computed: true},
					"status":      {Type: schema.TypeString, Computed: true},
				}},
			},
		},
	}
}

func dataSourceUCloudLBAttachmentsRead(d *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb attachment list, %s", err)
	}
	lbID := d.Get("load_balancer_id").(string)
	listenerID := d.Get("listener_id").(string)
	vserverSet, err := client.describeVServerById(lbID, listenerID)
	if err != nil {
		return fmt.Errorf("error on reading lb attachment list, %s", err)
	}
	var attachments []ulb.ULBBackendSet
	if ids, ok := d.GetOk("ids"); ok {
		for _, item := range vserverSet.BackendSet {
			if isStringIn(item.BackendId, schemaSetToStringSlice(ids)) {
				attachments = append(attachments, item)
			}
		}
	} else {
		attachments = vserverSet.BackendSet
	}
	if err := dataSourceUCloudLBAttachmentsSave(d, attachments); err != nil {
		return fmt.Errorf("error on reading lb attachment list, %s", err)
	}
	return nil
}

func dataSourceUCloudLBAttachmentsSave(d *schema.ResourceData, attachments []ulb.ULBBackendSet) error {
	ids := []string{}
	data := []map[string]interface{}{}
	for _, item := range attachments {
		ids = append(ids, item.BackendId)
		data = append(data, map[string]interface{}{
			"id": item.BackendId, "resource_id": item.ResourceId, "port": item.Port,
			"private_ip": item.PrivateIP, "status": lbAttachmentStatusCvt.convert(item.Status),
		})
	}
	d.SetId(hashStringArray(ids))
	_ = d.Set("total_count", len(data))
	_ = d.Set("ids", ids)
	if err := d.Set("lb_attachments", data); err != nil {
		return err
	}
	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}
	return nil
}
