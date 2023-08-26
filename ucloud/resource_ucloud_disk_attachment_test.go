package ucloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

func TestAccUCloudDiskAttachment_basic(t *testing.T) {
	var diskSet udisk.UDiskDataSet
	var instance uhost.UHostInstanceSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskAttachmentDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccDiskAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckDiskAttachmentExists("ucloud_disk_attachment.foo", &diskSet, &instance),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "20"),
				),
			},
			{
				Config: testAccDiskAttachmentConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckDiskAttachmentExists("ucloud_disk_attachment.foo", &diskSet, &instance),
					resource.TestCheckResourceAttr("ucloud_disk.foo", "disk_size", "40"),
				),
			},
		},
	})
}

func testAccCheckDiskAttachmentExists(n string, diskSet *udisk.UDiskDataSet, instance *uhost.UHostInstanceSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("disk attachment id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)

		diskId := rs.Primary.Attributes["disk_id"]
		resourceId := rs.Primary.Attributes["instance_id"]

		return resource.Retry(3*time.Minute, func() *resource.RetryError {
			d, err := client.describeDiskResource(diskId, resourceId)

			if err != nil {
				return resource.NonRetryableError(err)
			}

			if d.UDiskId == diskSet.UDiskId && d.UHostId == instance.UHostId {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("disk attachment not found"))
		})
	}
}

func testAccCheckDiskAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_disk_attachment" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeDiskById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.UHostId == rs.Primary.Attributes["instance_id"] {
			return fmt.Errorf("disk attachment still exists")
		}
	}

	return nil
}

const testAccDiskAttachmentConfig = `
variable "availability_zone" {
  type    = string
  default = "cn-bj2-05"
}

data "ucloud_images" "default" {
  availability_zone = "cn-bj2-05"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_disk" "foo" {
  availability_zone = "${var.availability_zone}"
  name              = "tf-acc-disk-attachment"
  disk_size         = 20
}

resource "ucloud_instance" "foo" {
  name                 = "tf-acc-disk-attachment"
  instance_type        = "n-highcpu-1"
  availability_zone    = "${var.availability_zone}"
  image_id             = "${data.ucloud_images.default.images.0.id}"
  charge_type          = "month"
  duration             = 1
  root_password        = "wA123456"
}

resource "ucloud_disk_attachment" "foo" {
  availability_zone = "${var.availability_zone}"
  disk_id           = "${ucloud_disk.foo.id}"
  instance_id       = "${ucloud_instance.foo.id}"
}
`

const testAccDiskAttachmentConfigUpdate = `
variable "availability_zone" {
  type    = string
  default = "cn-bj2-05"
}

data "ucloud_images" "default" {
  availability_zone = "${var.availability_zone}"
  name_regex        = "^CentOS 7.[1-2] 64"
  image_type        = "base"
}

resource "ucloud_disk" "foo" {
  availability_zone = "${var.availability_zone}"
  name              = "tf-acc-disk-attachment"
  disk_size         = 40
}

resource "ucloud_instance" "foo" {
  name                 = "tf-acc-disk-attachment"
  instance_type        = "n-highcpu-1"
  availability_zone    = "${var.availability_zone}"
  image_id             = "${data.ucloud_images.default.images.0.id}"
  charge_type          = "month"
  duration             = 1
  root_password        = "wA123456"
}

resource "ucloud_disk_attachment" "foo" {
  availability_zone              = "${var.availability_zone}"
  disk_id                        = "${ucloud_disk.foo.id}"
  instance_id                    = "${ucloud_instance.foo.id}"
  stop_instance_before_detaching = true
}
`
