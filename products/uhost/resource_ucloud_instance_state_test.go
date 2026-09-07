package uhost_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccUCloudInstanceState(t *testing.T) {
	instanceStateResource := "ucloud_instance_state.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: instanceStateResource,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceStateDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig(),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-basic"),
					resource.TestCheckResourceAttr(instanceStateResource, "state", "Stopped"),
				),
			},
			{
				Config: testAccInstanceStateConfigUpdate(),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-basic"),
					resource.TestCheckResourceAttr(instanceStateResource, "state", "Running"),
				),
			},
		},
	})
}

func testAccInstanceStateConfig() string {
	return fmt.Sprintf(`
variable "availability_zone" {
  type    = string
  default = "cn-bj2-05"
}
data "ucloud_images" "default" {
  availability_zone = "${var.availability_zone}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  availability_zone = "${var.availability_zone}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  charge_type       = "month"
  duration          = 0
  name              = "tf-acc-instance-config-basic"
  tag               = "tf-acc"
}
resource "ucloud_instance_state" "foo" {
	instance_id = "${ucloud_instance.foo.id}"
	force  = true
	state = "Stopped"
}
`)
}

func testAccInstanceStateConfigUpdate() string {
	return fmt.Sprintf(`
variable "availability_zone" {
  type    = string
  default = "cn-bj2-05"
}
data "ucloud_images" "default" {
  availability_zone = "${var.availability_zone}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_instance" "foo" {
  availability_zone = "${var.availability_zone}"
  image_id          = "${data.ucloud_images.default.images.0.id}"
  instance_type     = "n-highcpu-1"
  root_password     = "wA1234567"
  charge_type       = "month"
  duration          = 0
  name              = "tf-acc-instance-config-basic"
  tag               = "tf-acc"
}
resource "ucloud_instance_state" "foo" {
	instance_id = "${ucloud_instance.foo.id}"
	state = "Running"
}
`)
}
