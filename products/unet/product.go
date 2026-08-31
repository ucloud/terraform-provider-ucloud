package unet

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "unet"

type adapterV1 struct{}

var _ product.V1 = adapterV1{}

func New() product.V1 {
	return adapterV1{}
}

func (adapterV1) Registration() product.Registration {
	return product.Registration{
		Name: Name,
		Resources: map[string]*schema.Resource{
			"ucloud_eip":             resourceUCloudEIP(),
			"ucloud_eip_association": resourceUCloudEIPAssociation(),
			"ucloud_security_group":  resourceUCloudSecurityGroup(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_eips":            dataSourceUCloudEips(),
			"ucloud_security_groups": dataSourceUCloudSecurityGroups(),
		},
	}
}
