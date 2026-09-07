package uaccount

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

// Name is the stable product identity used by the provider runtime.
const Name = "uaccount"

type adapterV1 struct{}

var _ product.V1 = adapterV1{}

// New returns the stable product registration consumed by the provider.
func New() product.V1 {
	return adapterV1{}
}

func (adapterV1) Registration() product.Registration {
	return product.Registration{
		Name: Name,
		DataSources: map[string]*schema.Resource{
			"ucloud_projects": dataSourceUCloudProjects(),
			"ucloud_zones":    dataSourceUCloudZones(),
		},
	}
}
