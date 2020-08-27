package ucloud

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
)

func TestAccUCloudDBInstance_basic(t *testing.T) {
	var db udb.UDBInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_db_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDBInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists("ucloud_db_instance.foo", &db),
					testAccCheckDBInstanceAttributes(&db),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "name", "tf-acc-db-instance-basic"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_storage", "20"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_type", "mysql-ha-1"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine", "mysql"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine_version", "5.7"),
				),
			},

			{
				Config: testAccDBInstanceConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists("ucloud_db_instance.foo", &db),
					testAccCheckDBInstanceAttributes(&db),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "name", "tf-acc-db-instance-basic-update"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_storage", "30"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_type", "mysql-ha-2"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine", "mysql"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine_version", "5.7"),
				),
			},
		},
	})
}

func TestAccUCloudDBInstance_nvme(t *testing.T) {
	var db udb.UDBInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_db_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDBInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceNVMeConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists("ucloud_db_instance.foo", &db),
					testAccCheckDBInstanceAttributes(&db),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "name", "tf-acc-db-instance-basic"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_storage", "20"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_type", "mysql-ha-nvme-2"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine", "mysql"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine_version", "5.7"),
				),
			},
		},
	})
}

func TestAccUCloudDBInstance_parameter_group(t *testing.T) {
	var db udb.UDBInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_db_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDBInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfigParameterGroup,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists("ucloud_db_instance.foo", &db),
					testAccCheckDBInstanceAttributes(&db),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "name", "tf-acc-db-instance-parameter-group"),
				),
			},
		},
	})
}

func TestAccUCloudDBInstance_backup(t *testing.T) {
	var db udb.UDBInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_db_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDBInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDBInstanceConfigBackup,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists("ucloud_db_instance.foo", &db),
					testAccCheckDBInstanceAttributes(&db),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "name", "tf-acc-db-instance-backup"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_storage", "20"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_type", "mysql-ha-1"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine", "mysql"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine_version", "5.7"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_begin_time", "0"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_count", "6"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_black_list.#", "1"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_black_list."+strconv.Itoa(schema.HashString("test.%")), "test.%"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_date", "1111001"),
				),
			},

			{
				Config: testAccDBInstanceConfigBackupTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBInstanceExists("ucloud_db_instance.foo", &db),
					testAccCheckDBInstanceAttributes(&db),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "name", "tf-acc-db-instance-backup-update"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_storage", "20"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "instance_type", "mysql-ha-1"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine", "mysql"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "engine_version", "5.7"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_begin_time", "5"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_count", "6"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_black_list.#", "2"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_black_list."+strconv.Itoa(schema.HashString("test.%")), "test.%"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_black_list."+strconv.Itoa(schema.HashString("city.address")), "city.address"),
					resource.TestCheckResourceAttr("ucloud_db_instance.foo", "backup_date", "0001111"),
				),
			},
		},
	})
}

func testAccCheckDBInstanceExists(n string, db *udb.UDBInstanceSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("db instance id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeDBInstanceById(rs.Primary.ID)

		log.Printf("[INFO] db instance id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*db = *ptr
		return nil
	}
}

func testAccCheckDBInstanceAttributes(db *udb.UDBInstanceSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if db.DBId == "" {
			return fmt.Errorf("db instance id is empty")
		}
		return nil
	}
}

func testAccCheckDBInstanceDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_db_instance" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeDBInstanceById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.DBId != "" {
			return fmt.Errorf("db instance still exist")
		}
	}

	return nil
}

const testAccDBInstanceNVMeConfig = `
data "ucloud_zones" "default" {
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.3.id}"
	name 			  = "tf-acc-db-instance-basic"
	instance_storage  = 20
	instance_type	  = "mysql-ha-nvme-2"
	engine			  = "mysql"
	engine_version 	  = "5.7"
	password 		  = "2018_UClou"
}
`

const testAccDBInstanceConfig = `
data "ucloud_zones" "default" {
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name 			  = "tf-acc-db-instance-basic"
	instance_storage  = 20
	instance_type	  = "mysql-ha-1"
	engine			  = "mysql"
	engine_version 	  = "5.7"
	password 		  = "2018_UClou"
}
`

const testAccDBInstanceConfigTwo = `
data "ucloud_zones" "default" {
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name 			  = "tf-acc-db-instance-basic-update"
	instance_storage  = 30
	instance_type     = "mysql-ha-2"
	engine 			  = "mysql"
	engine_version    = "5.7"
	password		  = "2018_UClou"
}
`
const testAccDBInstanceConfigBackup = `
data "ucloud_zones" "default" {
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name 			  = "tf-acc-db-instance-backup"
	instance_storage  = 20
	instance_type 	  = "mysql-ha-1"
	engine 			  = "mysql"
	engine_version	  = "5.7"
	password 		  = "2018_UClou"
	backup_begin_time = 0
	backup_count	  = 6
	backup_black_list = ["test.%"]
	backup_date 	  = "1111001"
}
`
const testAccDBInstanceConfigBackupTwo = `
data "ucloud_zones" "default" {
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name 			  = "tf-acc-db-instance-backup-update"
	instance_storage  = 20
	instance_type 	  = "mysql-ha-1"
	engine 			  = "mysql"
	engine_version    = "5.7"
	password 		  = "2018_UClou"
	backup_begin_time = 5
	backup_count	  = 6
	backup_black_list = ["test.%", "city.address"]
	backup_date		  = "0001111"
}
`

const testAccDBInstanceConfigParameterGroup = `
data "ucloud_zones" "default" {
}

data "ucloud_db_parameter_groups" "default" {
	availability_zone = data.ucloud_zones.default.zones[0].id
	name_regex		  = "mysql5.7默认配置"
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name 			  = "tf-acc-db-instance-parameter-group"
	instance_storage  = 20
	instance_type	  = "mysql-ha-1"
	engine			  = "mysql"
	engine_version 	  = "5.7"
	password 		  = "2018_UClou"
    parameter_group   = "${data.ucloud_db_parameter_groups.default.parameter_groups.0.id}"
}
`
