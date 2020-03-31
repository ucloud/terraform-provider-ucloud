package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

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

func TestAccUCloudInstance_outstanding(t *testing.T) {
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
				Config: testAccInstanceConfigOutstanding,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-outstanding"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "o-standard-4"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "cpu", "4"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "memory", "16"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_localDisk(t *testing.T) {
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
				Config: testAccInstanceConfigLocalDisk,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-local-disk"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-basic-1"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "status", "Running"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "boot_disk_type", "local_normal"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "boot_disk_size", "40"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_isolationGroup(t *testing.T) {
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
				Config: testAccInstanceConfigIsolationGroup,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-isolation-group"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-basic-1"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_userData(t *testing.T) {
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
				Config: testAccInstanceConfigUserData,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-user-data"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-basic-1"),
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
  charge_type       = "month"
  duration          = 0
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
  availability_zone 		= "${data.ucloud_zones.default.zones.0.id}"
  image_id          		= "${data.ucloud_images.default.images.0.id}"
  security_group    		= "${ucloud_security_group.default.id}"
  instance_type     		= "n-basic-2"
  root_password     		= "wA1234567"
  charge_type       		= "month"
  duration          		= 0
  name              		= "tf-acc-instance-config-basic-update"
  tag               		= ""
  allow_stopping_for_update = true
}`, rInt)
}

const testAccInstanceConfigOutstanding = `
data "ucloud_images" "default" {
  availability_zone = "cn-bj2-05"
  name_regex        = "^高内核CentOS 7.6 64"
  image_type        = "base"
}
data "ucloud_security_groups" "default" {
	type = "recommend_web"
}

resource "ucloud_instance" "foo" {
  availability_zone = "cn-bj2-05"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  security_group    = "${data.ucloud_security_groups.default.security_groups.0.id}"
  instance_type     = "o-standard-4"
  boot_disk_type    = "cloud_ssd"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-config-outstanding"
  tag               = "tf-acc"
}
`

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
  availability_zone         = "${data.ucloud_zones.default.zones.0.id}"
  image_id                  = "${data.ucloud_images.default.images.0.id}"
  security_group            = "${ucloud_security_group.default.id}"
  instance_type             = "n-highcpu-1"
  root_password             = "wA1234567"
  name                      = "tf-acc-instance-size-update"
  tag                       = "tf-acc"
  boot_disk_size            = 30
  data_disk_size            = 30
  allow_stopping_for_update = true
  vpc_id                    = "${ucloud_vpc.default.id}"
  subnet_id                 = "${ucloud_subnet.default.id}"
}
`, rInt)
}

const testAccInstanceConfigLocalDisk = `
data "ucloud_zones" "default" {
}

data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  name              = "tf-acc-instance-local-disk"
  tag               = "tf-acc"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  boot_disk_size    = 40
  boot_disk_type    = "local_normal"
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
}
`

const testAccInstanceConfigIsolationGroup = `
data "ucloud_zones" "default" {
}

data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 6.5 64"
  image_type        = "base"
}

resource "ucloud_isolation_group" "default" {
	name = "tf-acc-instance-isolation-group"
}

resource "ucloud_instance" "foo" {
  name              = "tf-acc-instance-isolation-group"
  tag               = "tf-acc"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  isolation_group	=  "${ucloud_isolation_group.default.id}"
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
}
`

const testAccInstanceConfigUserData = `
data "ucloud_zones" "default" {
}

data "ucloud_security_groups" "default" {
  type = "recommend_web"
}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.6 64"
  image_type        = "base"
}


resource "ucloud_instance" "foo" {
  name              = "tf-acc-instance-user-data"
  tag               = "tf-acc"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  user_data		    = "I_am_user_data"
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
}
`
