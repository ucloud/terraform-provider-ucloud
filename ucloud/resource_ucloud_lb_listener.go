package ucloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/customdiff"
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

		CustomizeDiff: customdiff.All(
			customizeDiffLBMethodToListenType,
			customizeDiffLBProtocolToListenType,
		),

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
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"request_proxy",
					"packets_transmit",
				}, false),
			},

			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
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
	}
}

func resourceUCloudLBListenerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.ulbconn

	lbId := d.Get("load_balancer_id").(string)
	protocol := d.Get("protocol").(string)
	lbSet, err := client.describeLBById(lbId)
	if err != nil {
		return fmt.Errorf("error on reading lb %q when creating lb listener, %s", lbId, err)
	}

	req := conn.NewCreateVServerRequest()
	if v, ok := d.GetOk("listen_type"); ok {
		err := availableLBChoices.validate(lbSet.ULBType, protocol, v.(string))
		if err != nil {
			return err
		}
		req.ListenType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		if choices := availableLBChoices.availableChoices(lbSet.ULBType, protocol); len(choices) == 0 {
			return fmt.Errorf("The protocol can only be one of %q, %q when lb is intranet, got %q", "tcp", "udp", protocol)
		} else {
			req.ListenType = ucloud.String(upperCamelCvt.unconvert(choices[0]))
		}
	}

	req.ULBId = ucloud.String(lbId)
	req.Protocol = ucloud.String(upperCvt.unconvert(protocol))
	req.Method = ucloud.String(upperCamelCvt.unconvert(d.Get("method").(string)))

	if v, ok := d.GetOk("port"); ok {
		req.FrontendPort = ucloud.Int(v.(int))
	} else {
		switch protocol {
		case "http":
			req.FrontendPort = ucloud.Int(80)
		case "https":
			req.FrontendPort = ucloud.Int(443)
		default:
			req.FrontendPort = ucloud.Int(1024)
		}
	}

	if v, ok := d.GetOk("name"); ok {
		req.VServerName = ucloud.String(v.(string))
	} else {
		req.VServerName = ucloud.String(resource.PrefixedUniqueId("tf-lb-listener-"))
	}

	if v, ok := d.GetOkExists("idle_timeout"); ok {
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

	if vserverSet.MonitorType == lbMatchTypePath {
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
	if listenType == "request_proxy" && !isStringIn(method, []string{"roundrobin", "source", "weight_roundrobin", "leastconn"}) {
		return fmt.Errorf("the method can only be one of %q, %q, %q or %q when listen_type is %q",
			"roundrobin", "source", "weight_roundrobin", "leastconn", "request_proxy")
	}

	if listenType == "packets_transmit" && !isStringIn(method, []string{"consistent_hash", "source_port", "consistent_hash_port", "roundrobin", "source", "weight_roundrobin"}) {
		return fmt.Errorf("the method can only be one of %q, %q, %q, %q, %q or %q when listen_type is %q",
			"consistent_hash", "source_port", "consistent_hash_port", "roundrobin", "source", "weight_roundrobin", "packets_transmit")
	}

	return nil
}

func customizeDiffLBProtocolToListenType(diff *schema.ResourceDiff, v interface{}) error {
	listenType := diff.Get("listen_type").(string)
	protocol := diff.Get("protocol").(string)
	if listenType == "" {
		return nil
	}

	choices := map[string]struct{}{}
	for _, r := range availableLBChoices {
		if listenType != r.ListenType {
			continue
		}

		choices[r.Protocol] = struct{}{}
		if protocol == r.Protocol {
			return nil
		}
	}

	values := []string{}
	for k := range choices {
		values = append(values, k)
	}

	return fmt.Errorf("the protocol can only be one of %v, when listen_type is %q, got %q", values, listenType, protocol)
}

type lBChoice struct {
	Mode       string
	Protocol   string
	ListenType string
}

type lbChoices []lBChoice

var availableLBChoices = lbChoices{
	{"OuterMode", "http", "request_proxy"},
	{"OuterMode", "https", "request_proxy"},
	{"OuterMode", "tcp", "request_proxy"},
	{"OuterMode", "tcp", "packets_transmit"},
	{"OuterMode", "udp", "packets_transmit"},
	{"InnerMode", "tcp", "packets_transmit"},
	{"InnerMode", "udp", "packets_transmit"},
}

func (lc *lbChoices) validate(mode, protocol, listen_type string) error {
	choices := lc.availableChoices(mode, protocol)

	if listen_type != "" && !isStringIn(listen_type, choices) {
		if mode == "InnerMode" {
			return fmt.Errorf("the listen_type can only be one of %v, when protocol is %q in the intranet mode,  got %q", choices, protocol, listen_type)
		} else {
			return fmt.Errorf("the listen_type can only be one of %v, when protocol is %q in the extranet mode, got %q", choices, protocol, listen_type)
		}
	}

	return nil
}

func (lc *lbChoices) availableChoices(mode, protocol string) []string {
	choices := []string{}

	for _, r := range availableLBChoices {
		if mode == r.Mode && protocol == r.Protocol {
			choices = append(choices, r.ListenType)
		}
	}

	return choices
}
