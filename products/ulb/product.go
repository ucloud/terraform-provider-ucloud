package ulb

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "ulb"

type adapterV1 struct{}

var _ product.V1 = adapterV1{}

// New returns the stable product registration consumed by the provider.
func New() product.V1 {
	return adapterV1{}
}

func (adapterV1) Registration() product.Registration {
	return product.Registration{
		Name: Name,
		Resources: map[string]*schema.Resource{
			"ucloud_lb":                resourceUCloudLB(),
			"ucloud_lb_listener":       resourceUCloudLBListener(),
			"ucloud_lb_attachment":     resourceUCloudLBAttachment(),
			"ucloud_lb_rule":           resourceUCloudLBRule(),
			"ucloud_lb_ssl":            resourceUCloudLBSSL(),
			"ucloud_lb_ssl_attachment": resourceUCloudLBSSLAttachment(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_lbs":            dataSourceUCloudLBs(),
			"ucloud_lb_listeners":   dataSourceUCloudLBListeners(),
			"ucloud_lb_attachments": dataSourceUCloudLBAttachments(),
			"ucloud_lb_rules":       dataSourceUCloudLBRules(),
			"ucloud_lb_ssls":        dataSourceUCloudLBSSLs(),
		},
	}
}
