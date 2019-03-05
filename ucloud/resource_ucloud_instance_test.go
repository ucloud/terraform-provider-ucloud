package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

func TestAccUCloudInstance_basic(t *testing.T) {
	rInt := acctest.RandInt()
	var instance uhost.UHostInstanceSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigBasic(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-basic"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-highcpu-1"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "cpu", "1"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "memory", "1"),
				),
			},
			{
				Config: testAccInstanceConfigBasicUpdate(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-basic-update"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-basic-2"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "cpu", "2"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "memory", "4"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_vpc(t *testing.T) {
	rInt := acctest.RandInt()
	var instance uhost.UHostInstanceSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigVPC(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-vpc"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_size(t *testing.T) {
	rInt := acctest.RandInt()
	var instance uhost.UHostInstanceSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccInstancesConfigSize(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-size"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "data_disk_size", "20"),
				),
			},
			{
				Config: testAccInstancesConfigSizeUpdate(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-size-update"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "data_disk_size", "30"),
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
			return fmt.Errorf("instance still exist")
		}
	}

	return nil
}

func testAccInstanceConfigBasic(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-config-basic-%d"
  tag  = "tf-acc"

  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }
}

resource "ucloud_instance" "foo" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  security_group    = "${ucloud_security_group.default.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-config-basic"
  tag               = "tf-acc"
}`, rInt)
}

func testAccInstanceConfigBasicUpdate(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-config-basic-update-%d"
  tag  = ""

  rules {
    port_range = "20-80"
    protocol   = "tcp"
    cidr_block = "0.0.0.0/0"
  }
}

resource "ucloud_instance" "foo" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  security_group    = "${ucloud_security_group.default.id}"
  instance_type     = "n-basic-2"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-config-basic-update"
  tag               = ""
}`, rInt)
}

func testAccInstanceConfigVPC(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_vpc" "default" {
  name        = "tf-acc-instance-config-vpc"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
  name       = "tf-acc-instance-config-vpc"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-config-vpc-%d"
  tag  = "tf-acc"

  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }
}

resource "ucloud_instance" "foo" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  security_group    = "${ucloud_security_group.default.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-config-vpc"
  tag               = "tf-acc"
  vpc_id            = "${ucloud_vpc.default.id}"
  subnet_id         = "${ucloud_subnet.default.id}"
}`, rInt)
}

func testAccInstancesConfigSize(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_vpc" "default" {
  name        = "tf-acc-instance-size"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
  name       = "tf-acc-instance-size"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-size-%d"
  tag  = "tf-acc"

  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }
}

resource "ucloud_instance" "foo" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  security_group    = "${ucloud_security_group.default.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-size"
  tag               = "tf-acc"
  data_disk_size    = 20
  vpc_id            = "${ucloud_vpc.default.id}"
  subnet_id         = "${ucloud_subnet.default.id}"
}`, rInt)
}

func testAccInstancesConfigSizeUpdate(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_vpc" "default" {
  name        = "tf-acc-instance-size"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "default" {
  name       = "tf-acc-instance-size"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.default.id}"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-size-update-%d"
  tag  = "tf-acc"

  rules {
    port_range = "80"
    protocol   = "tcp"
    cidr_block = "192.168.0.0/16"
    policy     = "accept"
  }
}

resource "ucloud_instance" "foo" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  security_group    = "${ucloud_security_group.default.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-size-update"
  tag               = "tf-acc"
  boot_disk_size    = 30
  data_disk_size    = 30
  vpc_id            = "${ucloud_vpc.default.id}"
  subnet_id         = "${ucloud_subnet.default.id}"
}
`, rInt)
}
