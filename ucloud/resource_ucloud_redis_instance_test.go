package ucloud

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

func testAccCheckActiveStandbyRedisExists(n string, inst *umem.URedisGroupSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("active-standby redis id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeActiveStandbyRedisById(rs.Primary.ID)

		log.Printf("[INFO] active-standby redis id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*inst = *ptr
		return nil
	}
}

func testAccCheckActiveStandbyRedisDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_redis_instance" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeActiveStandbyRedisById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.GroupId != "" {
			return fmt.Errorf("active-standby redis still exist")
		}
	}

	return nil
}

func testAccCheckDistributedRedisExists(n string, inst *umem.UMemSpaceSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("distributed redis id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeDistributedRedisById(rs.Primary.ID)

		log.Printf("[INFO] distributed redis id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*inst = *ptr
		return nil
	}
}

func testAccCheckDistributedRedisDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_redis_instance" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeDistributedRedisById(rs.Primary.ID)

		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.SpaceId != "" {
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
