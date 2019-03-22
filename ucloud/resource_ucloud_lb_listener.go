package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudLBListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudLBListenerCreate,
		Update: resourceUCloudLBListenerUpdate,
		Read:   resourceUCloudLBListenerRead,
		Delete: resourceUCloudLBListenerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"http",
					"https",
					"tcp",
					"udp",
				}, false),
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"listen_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "request_proxy",
				ValidateFunc: validation.StringInSlice([]string{
					"request_proxy",
					"packets_transmit",
				}, false),
			},

			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      80,
				ValidateFunc: validation.IntBetween(1, 65535),
			},

			"idle_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 86400),
			},

			"method": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "roundrobin",
				ValidateFunc: validation.StringInSlice([]string{
					"roundrobin",
					"weight_roundrobin",
					"source",
					"source_port",
					"consistent_hash",
					"consistent_hash_port",
					"leastconn",
				}, false),
			},

			"persistence_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "none",
				ValidateFunc: validation.StringInSlice([]string{
					"server_insert",
					"user_defined",
					"none",
				}, false),
			},

			"persistence": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"health_check_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"port",
					"path",
				}, false),
			},

			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customizeDiffLBMethodToListenType,
	}
}

func resourceUCloudLBListenerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)

	req := conn.NewCreateVServerRequest()
	req.ULBId = ucloud.String(lbId)
	req.Protocol = ucloud.String(upperCvt.unconvert(d.Get("protocol").(string)))
	req.FrontendPort = ucloud.Int(d.Get("port").(int))
	req.ListenType = ucloud.String(upperCamelCvt.unconvert(d.Get("listen_type").(string)))
	req.Method = ucloud.String(upperCamelCvt.unconvert(d.Get("method").(string)))

	if v, ok := d.GetOk("name"); ok {
		req.VServerName = ucloud.String(v.(string))
	} else {
		req.VServerName = ucloud.String(resource.PrefixedUniqueId("tf-lb-listener-"))
	}

	if v, ok := d.GetOk("idle_timeout"); ok {
		req.ClientTimeout = ucloud.Int(v.(int))
	}

	if v, ok := d.GetOk("persistence_type"); ok {
		req.PersistenceType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	}

	if v, ok := d.GetOk("persistence"); ok {
		req.PersistenceInfo = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_type"); ok {
		checkType := v.(string)
		req.MonitorType = ucloud.String(upperCamelCvt.unconvert(checkType))
		if checkType == "path" {
			if v, ok := d.GetOk("domain"); ok {
				req.Domain = ucloud.String(v.(string))
			}

			if v, ok := d.GetOk("path"); ok {
				req.Path = ucloud.String(v.(string))
			}
		}
	}

	resp, err := conn.CreateVServer(req)
	if err != nil {
		return fmt.Errorf("error on creating lb listener, %s", err)
	}

	d.SetId(resp.VServerId)

	// after create lb listener, we need to wait it initialized
	stateConf := lbListenerWaitForState(client, lbId, d.Id())

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for lb listener %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudLBListenerRead(d, meta)
}

func resourceUCloudLBListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*UCloudClient).ulbconn

	d.Partial(true)

	isChanged := false
	req := conn.NewUpdateVServerAttributeRequest()
	req.ULBId = ucloud.String(d.Get("load_balancer_id").(string))
	req.VServerId = ucloud.String(d.Id())

	if d.HasChange("name") && !d.IsNewResource() {
		isChanged = true
		req.VServerName = ucloud.String(d.Get("name").(string))
	}

	if d.HasChange("protocol") && !d.IsNewResource() {
		isChanged = true
		req.Protocol = ucloud.String(upperCvt.unconvert(d.Get("protocol").(string)))
	}

	if d.HasChange("method") && !d.IsNewResource() {
		isChanged = true
		req.Method = ucloud.String(upperCamelCvt.unconvert(d.Get("method").(string)))
	}

	if d.HasChange("persistence_type") && !d.IsNewResource() {
		isChanged = true
		req.PersistenceType = ucloud.String(upperCamelCvt.unconvert(d.Get("persistence_type").(string)))
	}

	if d.HasChange("persistence") && !d.IsNewResource() {
		isChanged = true
		req.PersistenceInfo = ucloud.String(d.Get("persistence").(string))
	}

	if d.HasChange("idle_timeout") && !d.IsNewResource() {
		isChanged = true
		req.ClientTimeout = ucloud.Int(d.Get("idle_timeout").(int))
	}

	if d.HasChange("health_check_type") && !d.IsNewResource() {
		isChanged = true
		req.MonitorType = ucloud.String(upperCamelCvt.unconvert(d.Get("health_check_type").(string)))
	}

	if d.HasChange("domain") && !d.IsNewResource() {
		isChanged = true
		req.Domain = ucloud.String(d.Get("domain").(string))
	}

	if d.HasChange("path") && !d.IsNewResource() {
		isChanged = true
		req.Path = ucloud.String(d.Get("path").(string))
	}

	if isChanged {
		_, err := conn.UpdateVServerAttribute(req)
		if err != nil {
			return fmt.Errorf("error on %s to lb listener %q, %s", "UpdateVServerAttribute", d.Id(), err)
		}

		d.SetPartial("name")
		d.SetPartial("protocol")
		d.SetPartial("method")
		d.SetPartial("persistence_type")
		d.SetPartial("persistence")
		d.SetPartial("idle_timeout")
		d.SetPartial("health_check_type")
		d.SetPartial("domain")
		d.SetPartial("path")
	}

	d.Partial(false)

	return resourceUCloudLBListenerRead(d, meta)
}

func resourceUCloudLBListenerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	var err error
	var lbId string
	var vserverSet *ulb.ULBVServerSet

	if v, ok := d.GetOk("load_balancer_id"); ok {
		vserverSet, err = client.describeVServerById(v.(string), d.Id())
		if err != nil {
			if isNotFoundError(err) {
				d.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading lb listener %q, %s", d.Id(), err)
		}

		d.Set("load_balancer_id", v)
	} else {
		vserverSet, lbId, err = client.describeVServerByOneId(d.Id())
		if err != nil {
			return fmt.Errorf("error on parsing lb listener %q, %s", d.Id(), err)
		}

		d.Set("load_balancer_id", lbId)
	}

	d.Set("name", vserverSet.VServerName)
	d.Set("protocol", upperCvt.convert(vserverSet.Protocol))
	d.Set("listen_type", upperCamelCvt.convert(vserverSet.ListenType))
	d.Set("port", vserverSet.FrontendPort)
	d.Set("idle_timeout", vserverSet.ClientTimeout)
	d.Set("method", upperCamelCvt.convert(vserverSet.Method))
	d.Set("persistence_type", upperCamelCvt.convert(vserverSet.PersistenceType))
	d.Set("persistence", vserverSet.PersistenceInfo)
	d.Set("health_check_type", upperCamelCvt.convert(vserverSet.MonitorType))
	d.Set("status", listenerStatusCvt.convert(vserverSet.Status))

	if vserverSet.MonitorType == lbPath {
		d.Set("domain", vserverSet.Domain)
		d.Set("path", vserverSet.Path)
	}

	return nil
}

func resourceUCloudLBListenerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn
	lbId := d.Get("load_balancer_id").(string)

	req := conn.NewDeleteVServerRequest()
	req.ULBId = ucloud.String(lbId)
	req.VServerId = ucloud.String(d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := conn.DeleteVServer(req); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb listener %q, %s", d.Id(), err))
		}

		_, err := client.describeVServerById(lbId, d.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb listener when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified eip %q has not been deleted due to unknown error", d.Id()))
	})
}

func lbListenerWaitForState(client *UCloudClient, lbId, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			vserverSet, err := client.describeVServerById(lbId, id)
			if err != nil {
				if isNotFoundError(err) {
					return nil, statusPending, nil
				}
				return nil, "", err
			}

			return vserverSet, statusInitialized, nil
		},
	}
}

func customizeDiffLBMethodToListenType(diff *schema.ResourceDiff, v interface{}) error {
	listenType := diff.Get("listen_type").(string)
	method := diff.Get("method").(string)
	if listenType == "request_proxy" && (method == "source_port" || method == "consistent_hash" || method == "consistent_hash_port") {
		return fmt.Errorf("the method can only be one of %q, %q, %q or %q when listen_type is %q", "roundrobin", "source", "weight_roundrobin", "leastconn", "request_proxy")
	}

	if listenType == "packets_transmit" && (method == "roundrobin" || method == "weight_roundrobin" || method == "leastconn" || method == "source") {
		return fmt.Errorf("the method can only be one of %q, %q or %q when listen_type is %q", "source_port", "consistent_hash", "consistent_hash_port", "request_proxy")
	}

	return nil
}
