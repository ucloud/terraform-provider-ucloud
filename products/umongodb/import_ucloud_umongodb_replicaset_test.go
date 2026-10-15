package umongodb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudUMongoDBReplSet_import(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUMongoDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUMongoDBReplSetConfig,
			},
			{
				ResourceName:      "ucloud_umongodb_replicaset.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"admin_password",
					"charge_type",
					"quantity",
					"restart_trigger",
				},
			},
		},
	})
}
