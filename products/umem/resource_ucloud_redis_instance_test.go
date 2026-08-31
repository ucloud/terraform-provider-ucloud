package umem_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
)

func TestAccUCloudActiveStandbyRedis_basic(t *testing.T) {
	var inst umem.URedisGroupSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_redis_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckActiveStandbyRedisDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccActiveStandbyRedisConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "name", "tf-acc-redis"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "instance_type", "redis-master-1"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "engine_version", "4.0"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "backup_begin_time", "3"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "auto_backup", "disable"),
				),
			},

			{
				Config: testAccActiveStandbyRedisConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "name", "tf-acc-redis-renamed"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "instance_type", "redis-master-2"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "engine_version", "4.0"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "backup_begin_time", "0"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "auto_backup", "disable"),
				),
			},
		},
	})
}

func TestAccUCloudDistributedRedis_basic(t *testing.T) {
	var inst umem.UMemSpaceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_redis_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDistributedRedisDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDistributedRedisConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributedRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "instance_type", "redis-distributed-16"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "name", "tf-acc-redis"),
				),
			},

			{
				Config: testAccDistributedRedisConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributedRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "instance_type", "redis-distributed-20"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "name", "tf-acc-redis-renamed"),
				),
			},
		},
	})
}

func testAccCheckActiveStandbyRedisExists(name string, target *umem.URedisGroupSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("active-standby redis id is empty")
		}

		client, err := testAccUMemClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccActiveStandbyRedisByID(client, item.Primary.ID)
		log.Printf("[INFO] active-standby redis id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("active-standby redis %q is not found", item.Primary.ID)
		}
		*target = *instance
		return nil
	}
}

func testAccCheckActiveStandbyRedisDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_redis_instance" {
			continue
		}

		client, err := testAccUMemClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccActiveStandbyRedisByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if instance.GroupId != "" {
			return fmt.Errorf("active-standby redis still exist")
		}
	}
	return nil
}

func testAccCheckDistributedRedisExists(name string, target *umem.UMemSpaceSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("distributed redis id is empty")
		}

		client, err := testAccUMemClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccDistributedRedisByID(client, item.Primary.ID)
		log.Printf("[INFO] distributed redis id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("distributed redis %q is not found", item.Primary.ID)
		}
		*target = *instance
		return nil
	}
}

func testAccCheckDistributedRedisDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_redis_instance" {
			continue
		}

		client, err := testAccUMemClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccDistributedRedisByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if instance.SpaceId != "" {
			return fmt.Errorf("distributed redis still exist")
		}
	}
	return nil
}

const testAccActiveStandbyRedisConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-1"
	password = "2018_tfacc"
	name = "tf-acc-redis"
	tag = "tf-acc"
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
}
`

const testAccActiveStandbyRedisConfigUpdate = `
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-2"
	password = "2018_tfacc"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	backup_begin_time = 0
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
}
`

const testAccDistributedRedisConfig = `
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-redis"
	tag = "tf-acc"
	instance_type = "redis-distributed-16"
}
`

const testAccDistributedRedisConfigUpdate = `
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	instance_type = "redis-distributed-20"
}
`
