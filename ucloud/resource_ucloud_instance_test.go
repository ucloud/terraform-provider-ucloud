package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

func TestAccUCloudInstance_basic(t *testing.T) {
	var instance uhost.UHostInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigBasic,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-testAccInstanceConfigBasic"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-highcpu-1"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "cpu", "1"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "memory", "1024"),
				),
			},
			resource.TestStep{
				Config: testAccInstanceConfigBasicUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-testAccInstanceConfigBasicUpdate"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-basic-2"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "cpu", "2"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "memory", "4096"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_vpc(t *testing.T) {
	var instance uhost.UHostInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceConfigVPC,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-testAccInstanceConfigVPC"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_size(t *testing.T) {
	var instance uhost.UHostInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstancesConfigSize,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-testAccInstanceConfigSize"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "data_disk_size", "50"),
				),
			},
			resource.TestStep{
				Config: testAccInstancesConfigSizeUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-testAccInstanceConfigSize"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "data_disk_size", "100"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "boot_disk_size", "30"),
				),
			},
		},
	})
}

func testAccCheckInstanceExists(n string, instance *uhost.UHostInstanceSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("instance id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeInstanceById(rs.Primary.ID)

		log.Printf("[INFO] instance id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*instance = *ptr
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_instance" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		instance, err := client.describeInstanceById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if instance.State != "" && instance.State != string("Stopped") {
			return fmt.Errorf("found unstopped instance: %s", instance.UHostId)
		}

		if instance.UHostId != "" {
			return fmt.Errorf("UHostId still exist")
		}
	}

	return nil
}

const testAccInstanceConfigBasic = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	os_type = "Linux"
}

resource "ucloud_security_group" "default" {
    name = "tf-testAccInstanceConfigBasic"
    tag  = "tf-testInstance"

    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id = "${data.ucloud_images.default.images.0.id}"
	root_password = "wA1234567"
	security_group = "${ucloud_security_group.default.id}"
	name = "tf-testAccInstanceConfigBasic"
	instance_type = "n-highcpu-1"
	tag  = "tf-testInstance"
}
`

const testAccInstanceConfigBasicUpdate = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	os_type = "Linux"
}

resource "ucloud_security_group" "default" {
    name = "tf-testAccInstanceConfigBasicUpdate"
    tag  = "tf-testInstance"

	rules {
		port_range = "20-80"
		protocol   = "TCP"
		cidr_block = "0.0.0.0/0"
	}
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id = "${data.ucloud_images.default.images.0.id}"
	root_password = "wA1234567"
	security_group = "${ucloud_security_group.default.id}"
	name = "tf-testAccInstanceConfigBasicUpdate"
	instance_type = "n-basic-2"
	tag  = "tf-testInstanceTwo"
}
`
const testAccInstanceConfigVPC = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	os_type = "Linux"
}

resource "ucloud_vpc" "default" {
	name = "tf-testAccInstanceConfigVPC"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
	name = "tf-testAccInstanceConfigVPC"
	tag = "testTag"
	cidr_block = "192.168.1.0/24"
	vpc_id = "${ucloud_vpc.default.id}"
}

resource "ucloud_security_group" "default" {
    name = "tf-testAccInstanceConfigVPC"
    tag  = "tf-testInstance"

    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id = "${data.ucloud_images.default.images.0.id}"
	root_password = "wA1234567"
	security_group = "${ucloud_security_group.default.id}"
	name = "tf-testAccInstanceConfigVPC"
	instance_type = "n-highcpu-1"
	vpc_id    = "${ucloud_vpc.default.id}"
    subnet_id = "${ucloud_subnet.default.id}"
}
`
const testAccInstancesConfigSize = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	os_type = "Linux"
}

resource "ucloud_vpc" "default" {
	name = "tf-testAccInstanceConfigSize"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
	name = "tf-testAccInstanceConfigSize"
	tag = "testTag"
	cidr_block = "192.168.1.0/24"
	vpc_id = "${ucloud_vpc.default.id}"
}

resource "ucloud_security_group" "default" {
    name = "tf-testAccInstanceConfigSize"
    tag  = "tf-testInstance"

    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id = "${data.ucloud_images.default.images.0.id}"
	root_password = "wA1234567"
	security_group = "${ucloud_security_group.default.id}"
	name = "tf-testAccInstanceConfigSize"
	instance_type = "n-highcpu-1"
	vpc_id    = "${ucloud_vpc.default.id}"
	subnet_id = "${ucloud_subnet.default.id}"
	data_disk_size = 50
}
`
const testAccInstancesConfigSizeUpdate = `
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	os_type = "Linux"
}

resource "ucloud_vpc" "default" {
	name = "tf-testAccInstanceConfigSize"
	cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
	name = "tf-testAccInstanceConfigSize"
	tag = "testTag"
	cidr_block = "192.168.1.0/24"
	vpc_id = "${ucloud_vpc.default.id}"
}

resource "ucloud_security_group" "default" {
    name = "tf-testAccInstanceConfigSize"
    tag  = "tf-testInstance"

    rules {
        port_range = "80"
        protocol   = "TCP"
        cidr_block = "192.168.0.0/16"
        policy     = "ACCEPT"
    }
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id = "${data.ucloud_images.default.images.0.id}"
	root_password = "wA1234567"
	security_group = "${ucloud_security_group.default.id}"
	name = "tf-testAccInstanceConfigSize"
	instance_type = "n-highcpu-1"
	vpc_id    = "${ucloud_vpc.default.id}"
	subnet_id = "${ucloud_subnet.default.id}"
	boot_disk_size = 30
	data_disk_size = 100
}
`
