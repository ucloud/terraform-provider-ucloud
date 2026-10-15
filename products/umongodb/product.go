package umongodb

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "umongodb"

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
			"ucloud_umongodb_replicaset":      resourceUCloudUMongoDBReplSet(),
			"ucloud_umongodb_sharded_cluster": resourceUCloudUMongoDBShardedCluster(),
			"ucloud_umongodb_backup":          resourceUCloudUMongoDBBackup(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_umongodb_instance":  dataSourceUCloudUMongoDBInstance(),
			"ucloud_umongodb_instances": dataSourceUCloudUMongoDBInstances(),
		},
	}
}
