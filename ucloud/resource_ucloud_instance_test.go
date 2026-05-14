package ucloud

import (
	"fmt"
	"log"
	"os"
	"regexp"
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
					resource.TestCheckNoResourceAttr("ucloud_instance.foo", "key_pair_id"),
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

func TestAccUCloudInstance_passwordLoginMode(t *testing.T) {
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
				Config: testAccInstanceConfigPasswordLoginMode(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "login_mode", "Password"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "root_password", "wA1234567"),
					resource.TestCheckNoResourceAttr("ucloud_instance.foo", "key_pair_id"),
				),
			},
		},
	})
}

func TestAccUCloudInstance_keyPair(t *testing.T) {
	rInt := acctest.RandInt()
	keyPairID := os.Getenv("UCLOUD_KEY_PAIR_ID")
	if keyPairID == "" {
		t.Skip("UCLOUD_KEY_PAIR_ID must be set for key pair login acceptance tests")
	}

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
				Config: testAccInstanceConfigKeyPair(rInt, keyPairID),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "login_mode", "KeyPair"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "key_pair_id", keyPairID),
				),
			},
		},
	})
}

func TestAccUCloudInstance_invalidLoginMode(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfigKeyPairWithRootPassword,
				ExpectError: regexp.MustCompile(`(?i)key_pair_id|login_mode|root_password`),
			},
			{
				Config:      testAccInstanceConfigKeyPairWithoutKeyPairID,
				ExpectError: regexp.MustCompile(`(?i)key_pair_id|login_mode|KeyPair`),
			},
			{
				Config:      testAccInstanceConfigKeyPairIDWithoutLoginMode,
				ExpectError: regexp.MustCompile(`(?i)login_mode|key_pair_id`),
			},
			{
				Config:      testAccInstanceConfigPasswordWithKeyPairID,
				ExpectError: regexp.MustCompile(`(?i)key_pair_id|login_mode|Password`),
			},
			{
				Config:      testAccInstanceConfigUnsupportedLoginMode,
				ExpectError: regexp.MustCompile(`(?i)login_mode|Password|KeyPair`),
			},
		},
	})
}

func TestAccUCloudInstance_EIP(t *testing.T) {
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
				Config: testAccInstanceConfigEIP,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-eip"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
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

func TestAccUCloudInstance_dataDisks(t *testing.T) {
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
				Config: testAccInstanceConfigDataDisks,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-data-disks"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_instance.foo", "instance_type", "n-basic-1"),
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

func TestAccUCloudInstance_rdma(t *testing.T) {
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
				Config: testAccInstanceConfigRDMA,

				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					resource.TestCheckResourceAttrSet("ucloud_instance.foo", "rdma_cluster_id"),
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

func testAccInstanceConfigPasswordLoginMode(rInt int) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-password-login-%d"
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
  instance_type     = "n-basic-1"
  login_mode        = "Password"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-password-login"
  tag               = "tf-acc"
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
  boot_disk_type    = "cloud_rssd"
  root_password     = "wA1234567"
  name              = "tf-acc-instance-config-outstanding"
  tag               = "tf-acc"
  min_cpu_platform  = "Amd/Auto"
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

const testAccInstanceConfigDataDisks = `
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
  name              = "tf-acc-instance-data-disks"
  tag               = "tf-acc"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
  boot_disk_type = "cloud_ssd"
  data_disks {
    size = 20
    type = "cloud_ssd"
  }
  delete_disks_with_instance = true
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
  user_data		    = <<EOF
		#!/bin/bash
		sleep 5
		EOF
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
}
`

const testAccInstanceConfigEIP = `
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
  name              = "tf-acc-instance-eip"
  tag               = "tf-acc"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  security_group = "${data.ucloud_security_groups.default.security_groups.0.id}"
  network_interface {
	  eip_bandwidth     = 2
	  eip_charge_mode   = "bandwidth"
	  eip_internet_type = "bgp"
  }
  delete_eips_with_instance = true
}
`

const testAccInstanceConfigRDMA = `
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
  name              = "tf-acc-instance-eip"
  tag               = "tf-acc"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-basic-1"
  root_password     = "wA1234567"
  security_group    = "${data.ucloud_security_groups.default.security_groups.0.id}"
}
`

func testAccInstanceConfigKeyPair(rInt int, keyPairID string) string {
	return fmt.Sprintf(`
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_security_group" "default" {
  name = "tf-acc-instance-keypair-%d"
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
  instance_type     = "n-basic-1"
  login_mode        = "KeyPair"
  key_pair_id       = "%s"
  name              = "tf-acc-instance-keypair"
  tag               = "tf-acc"
}
`, rInt, keyPairID)
}

const testAccInstanceConfigInvalidLoginModeBase = `
resource "ucloud_instance" "foo" {
  availability_zone = "cn-bj2-02"
  image_id          = "uimage-test"
  security_group    = "secgroup-test"
  instance_type     = "n-basic-1"
  name              = "tf-acc-instance-invalid-login"
  tag               = "tf-acc"
`

const testAccInstanceConfigKeyPairWithRootPassword = testAccInstanceConfigInvalidLoginModeBase + `
  login_mode        = "KeyPair"
  key_pair_id       = "keypair-test"
  root_password     = "wA1234567"
}
`

const testAccInstanceConfigKeyPairWithoutKeyPairID = testAccInstanceConfigInvalidLoginModeBase + `
  login_mode        = "KeyPair"
}
`

const testAccInstanceConfigKeyPairIDWithoutLoginMode = testAccInstanceConfigInvalidLoginModeBase + `
  key_pair_id       = "keypair-test"
}
`

const testAccInstanceConfigPasswordWithKeyPairID = testAccInstanceConfigInvalidLoginModeBase + `
  login_mode        = "Password"
  key_pair_id       = "keypair-test"
  root_password     = "wA1234567"
}
`

const testAccInstanceConfigUnsupportedLoginMode = testAccInstanceConfigInvalidLoginModeBase + `
  login_mode        = "ImagePasswd"
  root_password     = "wA1234567"
}
`
