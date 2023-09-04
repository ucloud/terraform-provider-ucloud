package ucloud

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccUCloudBareMetalInstance_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBareMetalInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBareMetalInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ucloud_baremetal_instance.example", "availability_zone", "cn-bj2-02"),
					resource.TestCheckResourceAttr(
						"ucloud_baremetal_instance.example", "name", "UPHost"),
				),
			},
		},
	})
}

const testAccBareMetalInstanceConfig = `
resource "ucloud_baremetal_instance" "example" {
  availability_zone = "cn-bj2-02"
  image_id          = "pimg-cs-aqxttl"
  root_password     = "test123456"
  network_interface {
    eip_bandwidth = 10
    eip_charge_mode = "traffic"
    eip_internet_type = "bgp"
  }
  tag = "Default"
  instance_type     = "Base-SSD-V5"
  name              = "UPHost"
  raid_type         = "no_raid" 
  charge_type = "day"
  vpc_id = "uvnet-kq53a3nh"
  subnet_id = "subnet-vcqfv2tk"
  security_group = "firewall-rddft3jx"
}
`

func TestAccUCloudBareMetalInstance_cloud_disk(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBareMetalInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudDiskBareMetalInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ucloud_baremetal_instance.example", "availability_zone", "cn-sh2-01"),
					resource.TestCheckResourceAttr(
						"ucloud_baremetal_instance.example", "name", "UPHost"),
				),
			},
			{
				Config: testAccCloudDiskBareMetalInstanceUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"ucloud_baremetal_instance.example", "availability_zone", "cn-sh2-01"),
					resource.TestCheckResourceAttr(
						"ucloud_baremetal_instance.example", "name", "UPHost"),
				),
			},
		},
	})
}

const testAccCloudDiskBareMetalInstanceConfig = `
resource "ucloud_baremetal_instance" "example" {
  availability_zone = "cn-sh2-01"
  image_id          = "pimg-hd08-voxztd"
  root_password     = "test123456"
  boot_disk_size = "40"
  boot_disk_type = "cloud_rssd"
  data_disks {
    size = 40
    type = "cloud_rssd"
  }
  network_interface {
    eip_bandwidth = 10
    eip_charge_mode = "traffic"
    eip_internet_type = "bgp"
  }
  tag = "Default"
  instance_type     = "BM.Compute.I7.M10"
  name              = "UPHost"
  charge_type = "day"
  vpc_id = "uvnet-11gz43ik"
  subnet_id = "subnet-husd4aeb"
  security_group = "firewall-00ntra4z"
}
`

const testAccCloudDiskBareMetalInstanceUpdateConfig = `
resource "ucloud_baremetal_instance" "example" {
  availability_zone = "cn-sh2-01"
  image_id          = "pimg-hd08-voxztd"
  root_password     = "test123456"
  boot_disk_size = "100"
  boot_disk_type = "cloud_rssd"
  data_disks {
    size = 40
    type = "cloud_rssd"
  }
  network_interface {
    eip_bandwidth = 10
    eip_charge_mode = "traffic"
    eip_internet_type = "bgp"
  }

  tag = "Default"
  instance_type     = "BM.Compute.I7.M10"
  name              = "UPHost"
  charge_type = "day"
  vpc_id = "uvnet-11gz43ik"
  subnet_id = "subnet-husd4aeb"
  security_group = "firewall-3cmmotj3"
}
`

func testAccCheckBareMetalInstanceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*UCloudClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_baremetal_instance" {
			continue
		}

		_, err := client.describeBareMetalInstanceById(rs.Primary.ID)
		if isNotFoundError(err) {
			return nil
		}
		if err != nil {
			return err
		}
		return errors.New("instance still exists")
	}

	return nil
}
