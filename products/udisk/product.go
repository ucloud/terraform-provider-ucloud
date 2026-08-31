package udisk

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "udisk"

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
			"ucloud_disk":            resourceUCloudDisk(),
			"ucloud_disk_attachment": resourceUCloudDiskAttachment(),
			"ucloud_disk_snapshot":   resourceUCloudDiskSnapshot(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_disks":          dataSourceUCloudDisks(),
			"ucloud_disk_snapshots": dataSourceUCloudDiskSnapshots(),
		},
	}
}
