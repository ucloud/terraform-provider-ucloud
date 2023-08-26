package ucloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccUCloudInstanceState(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_instance_state.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceStateDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStateConfig(),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-basic"),
					resource.TestCheckResourceAttr("ucloud_instance_state.foo", "state", "Stopped"),
				),
			},
			{
				Config: testAccInstanceStateConfigUpdate(),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("ucloud_instance.foo", "name", "tf-acc-instance-config-basic"),
					resource.TestCheckResourceAttr("ucloud_instance_state.foo", "state", "Running"),
				),
			},
		},
	})
}

func testAccCheckInstanceStateDestroy(s *terraform.State) error {

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
