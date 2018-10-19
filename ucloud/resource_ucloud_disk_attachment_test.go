package ucloud

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
)

func TestAccUCloudDiskAttachment_basic(t *testing.T) {
	var diskSet udisk.UDiskDataSet
	var instance uhost.UHostInstanceSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_disk_attachment.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDiskAttachmentDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDiskAttachmentConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskExists("ucloud_disk.foo", &diskSet),
					testAccCheckInstanceExists("ucloud_instance.foo", &instance),
					testAccCheckDiskAttachmentExists("ucloud_disk_attachment.foo", &diskSet, &instance),
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
data "ucloud_zones" "default" {
}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex = "^CentOS 7.[1-2] 64"
	image_type =  "Base"
}

resource "ucloud_disk" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name = "testAcc"
	disk_size = 10
}

resource "ucloud_instance" "foo" {
	name = "testAccInstance"
	instance_type = "n-highcpu-1"
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id = "${data.ucloud_images.default.images.0.id}"
	instance_charge_type = "Month"
	instance_duration = 1
	root_password = "wA123456"
}

resource "ucloud_disk_attachment" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	disk_id = "${ucloud_disk.foo.id}"
	instance_id = "${ucloud_instance.foo.id}"
}
`
