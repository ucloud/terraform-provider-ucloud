package umem_test

import (
	"fmt"
	"log"
	"os"
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
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "status", "ISolation"),
				),
			},

			{
				Config: testAccActiveStandbyRedisConfigShutDown,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "status", "Running"),
				),
			},

			{
				// differs from the previous step only by restart_trigger, so this apply
				// exercises exactly the restart branch and nothing else; the instance
				// is Running here, which RestartURedisGroup requires
				Config: testAccActiveStandbyRedisConfigRestart,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "name", "tf-acc-redis-renamed"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "instance_type", "redis-master-2"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "restart_trigger", "tf-acc-restart-1"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "status", "Running"),
				),
			},

			{
				// differs from the previous step only by password, so this apply
				// exercises exactly the ModifyURedisGroupPassword branch and nothing
				// else; transform_type/restart_trigger keep their previous values so
				// neither re-fires. The instance is Running after the restart step,
				// which the password-change wait depends on.
				Config: testAccActiveStandbyRedisConfigPasswordUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckActiveStandbyRedisExists("ucloud_redis_instance.foo", &inst),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "password", "2019_tfacc"),
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "status", "Running"),
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
					resource.TestCheckResourceAttr("ucloud_redis_instance.foo", "instance_type", "redis-distributed-32"),
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

// The acceptance suite reuses one pre-created VPC/subnet (the UCloud region
// under test no longer sells classic-network UMem/URedis). Export UCLOUD_VPC_ID
// and UCLOUD_SUBNET_ID before running `make testacc PRODUCT=umem`; testAccPreCheck
// fails fast when either is missing.
var testAccActiveStandbyRedisConfig = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-1"
	password = "2018_tfacc"
	name = "tf-acc-redis"
	tag = "tf-acc"
	vpc_id = "%s"
	subnet_id = "%s"
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

var testAccActiveStandbyRedisConfigUpdate = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-2"
	password = "2018_tfacc"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	vpc_id = "%s"
	subnet_id = "%s"
	auto_backup = "disable"
	backup_begin_time = 0
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
	transform_type = "UNBind"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

var testAccActiveStandbyRedisConfigRestart = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-2"
	password = "2018_tfacc"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	vpc_id = "%s"
	subnet_id = "%s"
	auto_backup = "disable"
	backup_begin_time = 0
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
	transform_type = "Bind"
	restart_trigger = "tf-acc-restart-1"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

// testAccActiveStandbyRedisConfigPasswordUpdate differs from the restart config
// only by password, so its apply exercises exactly the password-change branch
// (ModifyURedisGroupPassword via the private SDK, which base64-encodes the
// password itself). Everything else — including transform_type and
// restart_trigger — keeps the previous step's value so no other branch fires.
var testAccActiveStandbyRedisConfigPasswordUpdate = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-2"
	password = "2019_tfacc"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	vpc_id = "%s"
	subnet_id = "%s"
	auto_backup = "disable"
	backup_begin_time = 0
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
	transform_type = "Bind"
	restart_trigger = "tf-acc-restart-1"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

// testAccActiveStandbyRedisConfigShutDown starts the instance back up
// (transform_type = "Bind") after the previous step isolated it. It must NOT
// set restart_trigger: the restart step below has to differ from this one only
// by restart_trigger, and a restart fired while the instance is still isolated
// is rejected by the backend (RetCode 21018).
var testAccActiveStandbyRedisConfigShutDown = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	engine_version = "4.0"
	instance_type = "redis-master-2"
	password = "2018_tfacc"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	vpc_id = "%s"
	subnet_id = "%s"
	auto_backup = "disable"
	backup_begin_time = 0
	standby_zone = "${data.ucloud_zones.default.zones.1.id}"
	transform_type = "Bind"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

var testAccDistributedRedisConfig = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-redis"
	tag = "tf-acc"
	instance_type = "redis-distributed-16"
	password = "2018_tfacc"
	block_cnt = 2
	vpc_id = "%s"
	subnet_id = "%s"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))

var testAccDistributedRedisConfigUpdate = fmt.Sprintf(`
data "ucloud_zones" "default" {}

resource "ucloud_redis_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "tf-acc-redis-renamed"
	tag = "tf-acc"
	instance_type = "redis-distributed-32"
	password = "9393_xnsjnj"
	block_cnt = 2
	vpc_id = "%s"
	subnet_id = "%s"
}
`, os.Getenv("UCLOUD_VPC_ID"), os.Getenv("UCLOUD_SUBNET_ID"))
