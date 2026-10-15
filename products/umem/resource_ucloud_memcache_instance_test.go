package umem_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
)

func TestAccUCloudActiveStandbyMemcache_basic(t *testing.T) {
	// Memcache is a legacy product that is no longer sold in most regions:
	// CreateUMemcacheGroup answers RetCode 150 "Service unavailable" (observed
	// in cn-wlcb). Skip by default; set UCLOUD_ACC_TEST_MEMCACHE=1 to run
	// against an account/region where the service is still available.
	if os.Getenv("UCLOUD_ACC_TEST_MEMCACHE") == "" {
		t.Skip("memcache service unavailable in most regions (RetCode 150); set UCLOUD_ACC_TEST_MEMCACHE=1 to force this test")
	}

	var inst pumem.UMemDataSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_memcache_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckActiveStandbyMemcacheDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccActiveStandbyMemcacheConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyMemcacheExists("ucloud_memcache_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_memcache_instance.foo", "name", "tf-acc-memcache"),
					resource.TestCheckResourceAttr("ucloud_memcache_instance.foo", "instance_type", "memcache-master-1"),
				),
			},

			{
				Config: testAccActiveStandbyMemcacheConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyMemcacheExists("ucloud_memcache_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_memcache_instance.foo", "name", "tf-acc-memcache-renamed"),
					resource.TestCheckResourceAttr("ucloud_memcache_instance.foo", "instance_type", "memcache-master-2"),
				),
			},
		},
	})
}

func testAccCheckActiveStandbyMemcacheExists(name string, target *pumem.UMemDataSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("active standby memcache id is empty")
		}

		client, err := testAccPrivateUMemClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccActiveStandbyMemcacheByID(client, item.Primary.ID)
		log.Printf("[INFO] active standby memcache id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("active standby memcache %q is not found", item.Primary.ID)
		}
		*target = *instance
		return nil
	}
}

func testAccCheckActiveStandbyMemcacheDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_memcache_instance" {
			continue
		}

		client, err := testAccPrivateUMemClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccActiveStandbyMemcacheByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if instance.ResourceId != "" {
			return fmt.Errorf("active standby memcache still exist")
		}
	}
	return nil
}

// See resource_ucloud_redis_instance_test.go: the suite reuses one pre-created
// VPC/subnet exported via UCLOUD_VPC_ID / UCLOUD_SUBNET_ID.
var testAccActiveStandbyMemcacheConfig = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_memcache_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-memcache"
	instance_type = "memcache-master-1"
	charge_type = "month"
	duration    = 1
	vpc_id = "%s"
	subnet_id = "%s"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

var testAccActiveStandbyMemcacheConfigUpdate = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_memcache_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-memcache-renamed"
	instance_type = "memcache-master-2"
	charge_type = "month"
	duration    = 1
	vpc_id = "%s"
	subnet_id = "%s"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))
