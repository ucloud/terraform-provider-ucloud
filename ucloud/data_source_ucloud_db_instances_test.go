package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudDBInstancesDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDBInstancesConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_db_instances.foo"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.0.name", fmt.Sprintf("tf-acc-db-instances-dataSource-basic-%d", rInt)),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.0.instance_type", "mysql-ha-1"),
				),
			},
		},
	})
}

func TestAccUCloudDBInstancesDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataDBInstancesConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_db_instances.foo"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.0.name", "tf-acc-db-instances-dataSource-ids"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_db_instances.foo", "db_instances.0.instance_type", "mysql-ha-1"),
				),
			},
		},
	})
}

func testAccDataDBInstancesConfig(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {
}

variable "name" {
	default = "tf-acc-db-instances-dataSource-basic-%d"
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name              = "${var.name}"
	instance_storage  = 20
	instance_type	  = "mysql-ha-1"
	engine			  = "mysql"
	engine_version 	  = "5.7"
	password 		  = "2018_UClou"
    tag               = "tf-acc"
}

data "ucloud_db_instances" "foo" {
	name_regex  = "${ucloud_db_instance.foo.name}"
}
`, rInt)
}

const testAccDataDBInstancesConfigIds = `
data "ucloud_zones" "default" {
}

variable "name" {
	default = "tf-acc-db-instances-dataSource-ids"
}

resource "ucloud_db_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name 			  = "${var.name}"
	instance_storage  = 20
	instance_type	  = "mysql-ha-1"
	engine			  = "mysql"
	engine_version 	  = "5.7"
	password 		  = "2018_UClou"
	tag               = "tf-acc"
}

data "ucloud_db_instances" "foo" {
	ids = ["${ucloud_db_instance.foo.id}"]
}
`
