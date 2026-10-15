package umongodb_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func TestAccUCloudUMongoDBReplSet_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUMongoDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUMongoDBReplSetConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_umongodb_replicaset.foo", "name", "tf-acc-umongodb-rs"),
					resource.TestCheckResourceAttr("ucloud_umongodb_replicaset.foo", "node_count", "3"),
					resource.TestCheckResourceAttr("ucloud_umongodb_replicaset.foo", "status", "running"),
				),
			},
		},
	})
}

func TestAccUCloudUMongoDBShardedCluster_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUMongoDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUMongoDBShardedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_umongodb_sharded_cluster.foo", "name", "tf-acc-umongodb-sc"),
					resource.TestCheckResourceAttr("ucloud_umongodb_sharded_cluster.foo", "shard_count", "2"),
				),
			},
		},
	})
}

func testAccCheckUMongoDBClusterDestroy(state *terraform.State) error {
	client, err := testAccUMongoDBClient()
	if err != nil {
		return err
	}
	for _, resourceState := range state.RootModule().Resources {
		switch resourceState.Type {
		case "ucloud_umongodb_replicaset", "ucloud_umongodb_sharded_cluster":
		default:
			continue
		}
		req := client.NewDescribeUMongoDBInstanceRequest()
		req.ClusterId = ucloud.String(resourceState.Primary.ID)
		resp, err := client.DescribeUMongoDBInstance(req)
		if err != nil {
			return fmt.Errorf("describe %s while checking destroy: %w", resourceState.Primary.ID, err)
		}
		if resp != nil && resp.ClusterInfo.ClusterId != "" && resp.ClusterInfo.State != "Deleted" {
			return fmt.Errorf("umongodb cluster %s still exists", resourceState.Primary.ID)
		}
	}
	return nil
}

const testAccUMongoDBReplSetConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_umongodb_replicaset" "foo" {
  availability_zone = data.ucloud_zones.default.zones[0].id
  name              = "tf-acc-umongodb-rs"
  db_version        = "MongoDB 4.2"
  admin_password    = "Tfacc123456"
  disk_space        = 20
  machine_type_id   = "o.mongo2m.medium"
  node_count        = 3
}
`

const testAccUMongoDBShardedConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_umongodb_sharded_cluster" "foo" {
  availability_zone  = data.ucloud_zones.default.zones[0].id
  name               = "tf-acc-umongodb-sc"
  db_version         = "MongoDB 4.2"
  admin_password     = "Tfacc123456"
  disk_space         = 20
  machine_type_id    = "o.mongo2m.medium"
  node_count         = 3
  shard_count        = 2
  mongos_node_count  = 2
}
`
