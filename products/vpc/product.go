package vpc

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "vpc"

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
			"ucloud_vpc":                    resourceUCloudVPC(),
			"ucloud_subnet":                 resourceUCloudSubnet(),
			"ucloud_vpc_peering_connection": resourceUCloudVPCPeeringConnection(),
			"ucloud_vip":                    resourceUCloudVIP(),
			"ucloud_nat_gateway":            resourceUCloudNatGateway(),
			"ucloud_nat_gateway_rule":       resourceUCloudNatGatewayRule(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_vpcs":         dataSourceUCloudVPCs(),
			"ucloud_subnets":      dataSourceUCloudSubnets(),
			"ucloud_nat_gateways": dataSourceUCloudNatGateways(),
			"ucloud_sec_groups":   dataSourceUCloudSecGroups(),
		},
	}
}
