package ucloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccUCloudSecurityGroupsDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSecurityGroupsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_security_groups.foo"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.type", "user defined"),
				),
			},
		},
	})
}

func TestAccUCloudSecurityGroupsDataSource_ids(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSecurityGroupsConfigIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_security_groups.foo"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.type", "user defined"),
				),
			},
		},
	})
}

func TestAccUCloudSecurityGroupsDataSource_security_resourceId(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSecurityGroupsConfigResourceId(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.tag", "tf-acc"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.type", "user defined"),
				),
			},
		},
	})
}

func testAccDataSecurityGroupsConfig(rInt int) string {
	return fmt.Sprintf(`
variable "name" {
	default = "tf-acc-sgs-basic"
}
resource "ucloud_security_group" "foo" {
	name = "${var.name}-%d"
	tag  = "tf-acc"
	rules {
		port_range = "80"
		protocol   = "tcp"
		cidr_block = "192.168.0.0/16"
		policy     = "accept"
		priority   = "high"
	}
}

data "ucloud_security_groups" "foo" {
	name_regex  = "${ucloud_security_group.foo.name}"
}
`, rInt)
}

func testAccDataSecurityGroupsConfigIds(rInt int) string {
	return fmt.Sprintf(`
variable "name" {
	default = "tf-acc-sgs-ids"
}
resource "ucloud_security_group" "foo" {
	name = "${var.name}-%d"
	tag  = "tf-acc"
	rules {
		port_range = "80"
		protocol   = "tcp"
		cidr_block = "192.168.0.0/16"
		policy     = "accept"
		priority   = "high"
	}
}

data "ucloud_security_groups" "foo" {
	ids = ["${ucloud_security_group.foo.*.id}"]
}
`, rInt)
}

func testAccDataSecurityGroupsConfigResourceId(rInt int) string {
	return fmt.Sprintf(`
variable "name" {
	default = "tf-acc-sgs-resourceId"
}

data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        = "base"
}  

resource "ucloud_security_group" "foo" {
	name = "${var.name}-%d"
	tag  = "tf-acc"
	rules {
		port_range = "80"
		protocol   = "tcp"
		cidr_block = "192.168.0.0/16"
		policy     = "accept"
		priority   = "high"
	}
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	security_group    = "${ucloud_security_group.foo.id}"
	instance_type     = "n-highcpu-1"
	root_password     = "wA1234567"
	name              = "${var.name}"
	tag               = "tf-acc"
}

data "ucloud_security_groups" "foo" {
	resource_id = "${ucloud_instance.foo.id}"
}
`, rInt)
}
