package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
)

func TestAccUCloudActiveStandbyMemcache_basic(t *testing.T) {
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

func testAccCheckActiveStandbyMemcacheExists(n string, inst *pumem.UMemDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("active standby memcache id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeActiveStandbyMemcacheById(rs.Primary.ID)

		log.Printf("[INFO] active standby memcache id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*inst = *ptr
		return nil
	}
}

func testAccCheckActiveStandbyMemcacheDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_memcache_instance" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeActiveStandbyMemcacheById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.ResourceId != "" {
			return fmt.Errorf("active standby memcache still exist")
		}
	}

	return nil
}

const testAccActiveStandbyMemcacheConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_memcache_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-memcache"
	instance_type = "memcache-master-1"
    charge_type = "month"
    duration    = 1
}
`

const testAccActiveStandbyMemcacheConfigUpdate = `
data "ucloud_zones" "default" {}

resource "ucloud_memcache_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-memcache-renamed"
	instance_type = "memcache-master-2"
    charge_type = "month"
    duration    = 1
}
`
