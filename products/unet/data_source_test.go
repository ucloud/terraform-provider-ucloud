package unet_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudEipsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataEipsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_eips.foo"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.name", "tf-acc-eips-dataSource-basic"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.share_bandwidth_package_id", ""),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.share_bandwidth_package_name", ""),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.share_bandwidth", "0"),
				),
			},
		},
	})
}

func TestAccUCloudEipsDataSource_ids(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataEipsConfigIds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_eips.foo"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.#", "2"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.0.bandwidth", "1"),
					resource.TestCheckResourceAttr("data.ucloud_eips.foo", "eips.1.charge_type", "month"),
				),
			},
		},
	})
}

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
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.rules.0.port_range", "80"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.type", "user_defined"),
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
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.rules.0.port_range", "80"),
				),
			},
		},
	})
}

func TestAccUCloudSecurityGroupsDataSource_type(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSecurityGroupsConfigType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDExists("data.ucloud_security_groups.foo"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("data.ucloud_security_groups.foo", "security_groups.0.type", "recommend_non_web"),
				),
			},
		},
	})
}

const testAccDataEipsConfig = `
variable "name" {
	default = "tf-acc-eips-dataSource-basic"
}

resource "ucloud_eip" "foo" {
	name          = "${var.name}"
	bandwidth     = 1
	internet_type = "bgp"
	duration      = 1
}

data "ucloud_eips" "foo" {
	name_regex = "${ucloud_eip.foo.name}"
}
`

const testAccDataEipsConfigIds = `
variable "name" {
	default = "tf-acc-eips-dataSource-ids"
}

variable "instance_count" {
	default = 2
}

variable "count_format" {
	default = "%02d"
}

resource "ucloud_eip" "foo" {
	name          = "${var.name}-${format(var.count_format, count.index+1)}"
	bandwidth     = 1
	internet_type = "bgp"
	duration      = 1
	count         = var.instance_count
}

data "ucloud_eips" "foo" {
	ids = ucloud_eip.foo.*.id
}
`

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
	name_regex = "${ucloud_security_group.foo.name}"
	type       = "user_defined"
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
	ids = ucloud_security_group.foo.*.id
}
`, rInt)
}

const testAccDataSecurityGroupsConfigType = `
data "ucloud_security_groups" "foo" {
	type = "recommend_non_web"
}
`
