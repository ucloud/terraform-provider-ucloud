package ulb

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

func resourceUCloudLBListenerCreate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when creating lb listener, %s", err)
	}
	conn := client.ulbconn

	lbID := data.Get("load_balancer_id").(string)
	protocol := data.Get("protocol").(string)
	lbSet, err := client.describeLBById(lbID)
	if err != nil {
		return fmt.Errorf("error on reading lb %q when creating lb listener, %s", lbID, err)
	}

	request := conn.NewCreateVServerRequest()
	if value, ok := data.GetOk("listen_type"); ok {
		lbListenType := upperCamelConvert(lbSet.ListenType)
		if isStringIn(lbListenType, []string{"request_proxy", "packets_transmit"}) && lbListenType != value.(string) {
			return fmt.Errorf("the %q of lb listener must be same as the lb's %q, got %q", "listen_type", lbListenType, value.(string))
		}
		if err := availableLBChoices.validate(protocol, value.(string)); err != nil {
			return err
		}
		request.ListenType = ucloud.String(upperCamelUnconvert(value.(string)))
	} else if choices := availableLBChoices.availableChoices(protocol); len(choices) == 0 {
		return fmt.Errorf("the protocol %q does not have available listen_type", protocol)
	} else {
		lbListenType := upperCamelConvert(lbSet.ListenType)
		if isStringIn(lbListenType, []string{"request_proxy", "packets_transmit"}) {
			request.ListenType = ucloud.String(lbSet.ListenType)
		} else {
			request.ListenType = ucloud.String(upperCamelUnconvert(choices[0]))
		}
	}

	request.ULBId = ucloud.String(lbID)
	request.Protocol = ucloud.String(upperUnconvert(protocol))
	request.Method = ucloud.String(upperCamelUnconvert(data.Get("method").(string)))

	if value, ok := data.GetOk("port"); ok {
		request.FrontendPort = ucloud.Int(value.(int))
	} else {
		switch protocol {
		case "http":
			request.FrontendPort = ucloud.Int(80)
		case "https":
			request.FrontendPort = ucloud.Int(443)
		default:
			request.FrontendPort = ucloud.Int(1024)
		}
	}
	if value, ok := data.GetOk("name"); ok {
		request.VServerName = ucloud.String(value.(string))
	} else {
		request.VServerName = ucloud.String(resource.PrefixedUniqueId("tf-lb-listener-"))
	}
	if value, ok := data.GetOkExists("idle_timeout"); ok {
		request.ClientTimeout = ucloud.Int(value.(int))
	}
	if value, ok := data.GetOk("persistence_type"); ok {
		request.PersistenceType = ucloud.String(upperCamelUnconvert(value.(string)))
	}
	if value, ok := data.GetOk("persistence"); ok {
		request.PersistenceInfo = ucloud.String(value.(string))
	}
	if value, ok := data.GetOk("health_check_type"); ok {
		checkType := value.(string)
		request.MonitorType = ucloud.String(upperCamelUnconvert(checkType))
		if checkType == "path" {
			if value, ok := data.GetOk("domain"); ok {
				request.Domain = ucloud.String(value.(string))
			}
			if value, ok := data.GetOk("path"); ok {
				request.Path = ucloud.String(value.(string))
			}
		}
	}

	response, err := conn.CreateVServer(request)
	if err != nil {
		return fmt.Errorf("error on creating lb listener, %s", err)
	}
	data.SetId(response.VServerId)

	if _, err = lbListenerWaitForState(client, lbID, data.Id()).WaitForState(); err != nil {
		return fmt.Errorf("error on waiting for lb listener %q complete creating, %s", data.Id(), err)
	}
	return resourceUCloudLBListenerRead(data, meta)
}

func resourceUCloudLBListenerUpdate(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when updating lb listener, %s", err)
	}
	conn := client.ulbconn

	data.Partial(true)
	isChanged := false
	request := conn.NewUpdateVServerAttributeRequest()
	request.ULBId = ucloud.String(data.Get("load_balancer_id").(string))
	request.VServerId = ucloud.String(data.Id())
	if data.HasChange("name") && !data.IsNewResource() {
		isChanged = true
		request.VServerName = ucloud.String(data.Get("name").(string))
	}
	if data.HasChange("method") && !data.IsNewResource() {
		isChanged = true
		request.Method = ucloud.String(upperCamelUnconvert(data.Get("method").(string)))
	}
	if data.HasChange("persistence_type") && !data.IsNewResource() {
		isChanged = true
		request.PersistenceType = ucloud.String(upperCamelUnconvert(data.Get("persistence_type").(string)))
	}
	if data.HasChange("persistence") && !data.IsNewResource() {
		isChanged = true
		request.PersistenceInfo = ucloud.String(data.Get("persistence").(string))
	}
	if data.HasChange("idle_timeout") && !data.IsNewResource() {
		isChanged = true
		request.ClientTimeout = ucloud.Int(data.Get("idle_timeout").(int))
	}
	if data.HasChange("health_check_type") && !data.IsNewResource() {
		isChanged = true
		request.MonitorType = ucloud.String(upperCamelUnconvert(data.Get("health_check_type").(string)))
	}
	if data.HasChange("domain") && !data.IsNewResource() {
		isChanged = true
		request.Domain = ucloud.String(data.Get("domain").(string))
	}
	if data.HasChange("path") && !data.IsNewResource() {
		isChanged = true
		request.Path = ucloud.String(data.Get("path").(string))
	}
	if isChanged {
		if _, err := conn.UpdateVServerAttribute(request); err != nil {
			return fmt.Errorf("error on %s to lb listener %q, %s", "UpdateVServerAttribute", data.Id(), err)
		}
		data.SetPartial("name")
		data.SetPartial("method")
		data.SetPartial("persistence_type")
		data.SetPartial("persistence")
		data.SetPartial("idle_timeout")
		data.SetPartial("health_check_type")
		data.SetPartial("domain")
		data.SetPartial("path")
	}
	data.Partial(false)
	return resourceUCloudLBListenerRead(data, meta)
}

func resourceUCloudLBListenerRead(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when reading lb listener, %s", err)
	}
	var lbID string
	var vserverSet *ulb.ULBVServerSet
	if value, ok := data.GetOk("load_balancer_id"); ok {
		vserverSet, err = client.describeVServerById(value.(string), data.Id())
		if err != nil {
			if isNotFoundError(err) {
				data.SetId("")
				return nil
			}
			return fmt.Errorf("error on reading lb listener %q, %s", data.Id(), err)
		}
		_ = data.Set("load_balancer_id", value)
	} else {
		vserverSet, lbID, err = client.describeVServerByOneId(data.Id())
		if err != nil {
			return fmt.Errorf("error on parsing lb listener %q, %s", data.Id(), err)
		}
		_ = data.Set("load_balancer_id", lbID)
	}

	_ = data.Set("name", vserverSet.VServerName)
	_ = data.Set("protocol", upperConvert(vserverSet.Protocol))
	_ = data.Set("listen_type", upperCamelConvert(vserverSet.ListenType))
	_ = data.Set("port", vserverSet.FrontendPort)
	_ = data.Set("idle_timeout", vserverSet.ClientTimeout)
	_ = data.Set("method", upperCamelConvert(vserverSet.Method))
	_ = data.Set("persistence_type", upperCamelConvert(vserverSet.PersistenceType))
	_ = data.Set("persistence", vserverSet.PersistenceInfo)
	_ = data.Set("health_check_type", upperCamelConvert(vserverSet.MonitorType))
	_ = data.Set("status", listenerStatusCvt.convert(vserverSet.Status))
	if vserverSet.MonitorType == lbMatchTypePath {
		_ = data.Set("domain", vserverSet.Domain)
		_ = data.Set("path", vserverSet.Path)
	}
	return nil
}

func resourceUCloudLBListenerDelete(data *schema.ResourceData, meta interface{}) error {
	client, err := clientFromMeta(meta)
	if err != nil {
		return fmt.Errorf("error on getting client when deleting lb listener, %s", err)
	}
	lbID := data.Get("load_balancer_id").(string)
	request := client.ulbconn.NewDeleteVServerRequest()
	request.ULBId = ucloud.String(lbID)
	request.VServerId = ucloud.String(data.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.ulbconn.DeleteVServer(request); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error on deleting lb listener %q, %s", data.Id(), err))
		}
		_, err := client.describeVServerById(lbID, data.Id())
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error on reading lb listener when deleting %q, %s", data.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("the specified eip %q has not been deleted due to unknown error", data.Id()))
	})
}

func lbListenerWaitForState(client *productClient, lbID, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusInitialized},
		Timeout:    3 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			vserverSet, err := client.describeVServerById(lbID, id)
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

func customizeDiffLBMethodToListenType(diff *schema.ResourceDiff, _ interface{}) error {
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

func customizeDiffLBProtocolToListenType(diff *schema.ResourceDiff, _ interface{}) error {
	listenType := diff.Get("listen_type").(string)
	protocol := diff.Get("protocol").(string)
	if listenType == "" {
		return nil
	}
	choices := map[string]struct{}{}
	for _, choice := range availableLBChoices {
		if listenType != choice.ListenType {
			continue
		}
		choices[choice.Protocol] = struct{}{}
		if protocol == choice.Protocol {
			return nil
		}
	}
	values := []string{}
	for value := range choices {
		values = append(values, value)
	}
	return fmt.Errorf("the protocol can only be one of %v, when listen_type is %q, got %q", values, listenType, protocol)
}

type lBChoice struct {
	Protocol   string
	ListenType string
}

type lbChoices []lBChoice

var availableLBChoices = lbChoices{
	{"http", "request_proxy"},
	{"https", "request_proxy"},
	{"tcp", "request_proxy"},
	{"tcp", "packets_transmit"},
	{"udp", "request_proxy"},
	{"udp", "packets_transmit"},
}

func (choices *lbChoices) validate(protocol, listenType string) error {
	available := choices.availableChoices(protocol)
	if listenType != "" && !isStringIn(listenType, available) {
		return fmt.Errorf("the listen_type can only be one of %v, when protocol is %q, got %q", available, protocol, listenType)
	}
	return nil
}

func (choices *lbChoices) availableChoices(protocol string) []string {
	result := []string{}
	for _, choice := range availableLBChoices {
		if protocol == choice.Protocol {
			result = append(result, choice.ListenType)
		}
	}
	return result
}
