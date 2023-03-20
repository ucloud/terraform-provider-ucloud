package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uads"
)

func TestAccUCloudAntiDDoS_basic(t *testing.T) {
	var uadsServiceInfo uads.ServiceInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_anti_ddos_instance.foo",
		Providers:     testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testAccAntiDDoSInstanceConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckAntiDDoSInstanceExists("ucloud_anti_ddos_instance.foo", &uadsServiceInfo),
					testAccCheckAntiDDoSInstanceAttributes(&uadsServiceInfo),
					resource.TestCheckResourceAttr("ucloud_anti_ddos_instance.foo", "name", "tf-acc-anti-ddos-instance-basic"),
					resource.TestCheckResourceAttr("ucloud_anti_ddos_instance.foo", "bandwidth", "50"),
				),
			},

			{
				Config: testAccAntiDDoSInstanceConfigUpdate,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckAntiDDoSInstanceExists("ucloud_anti_ddos_instance.foo", &uadsServiceInfo),
					testAccCheckAntiDDoSInstanceAttributes(&uadsServiceInfo),
					resource.TestCheckResourceAttr("ucloud_anti_ddos_instance.foo", "name", "tf-acc-anti-ddos-instance-basic"),
					resource.TestCheckResourceAttr("ucloud_anti_ddos_instance.foo", "bandwidth", "80"),
				),
			},
		},
	})
}

func testAccCheckAntiDDoSInstanceExists(n string, uadsServiceInfo *uads.ServiceInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("uads id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUADSById(rs.Primary.ID)

		log.Printf("[INFO] disk id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*uadsServiceInfo = *ptr
		return nil
	}
}

func testAccCheckAntiDDoSInstanceAttributes(uadsServiceInfo *uads.ServiceInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if uadsServiceInfo.ResourceId == "" {
			return fmt.Errorf("uads id is empty")
		}
		return nil
	}
}

func testAccCheckAntiDDoSInstanceDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_anti_ddos_instance" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUADSById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.ResourceId != "" {
			return fmt.Errorf("ucloud_anti_ddos_instance still exist")
		}
	}

	return nil
}

const testAccAntiDDoSInstanceConfig = `
resource "ucloud_anti_ddos_instance" "foo" {
    area               = "EastChina"
    bandwidth          = 50
    base_defence_value = 30
    data_center        = "Zaozhuang"
    max_defence_value  = 30
    name               = "tf-acc-anti-ddos-instance-basic"
}
`

const testAccAntiDDoSInstanceConfigUpdate = `
resource "ucloud_anti_ddos_instance" "foo" {
    area               = "EastChina"
    bandwidth          = 80
    base_defence_value = 30
    data_center        = "Zaozhuang"
    max_defence_value  = 30
    name               = "tf-acc-anti-ddos-instance-basic"
}
resource "ucloud_anti_ddos_allowed_domain" "foo" {
    domain      = "ucloud.cn"
    instance_id = "${ucloud_anti_ddos_instance.foo.id}"
    comment = "test-acc-comment"
}
resource "ucloud_anti_ddos_ip" "foo" {
    instance_id = "${ucloud_anti_ddos_instance.foo.id}"
    comment = "test-acc-comment"
}
resource "ucloud_anti_ddos_rule" "foo" {
    instance_id = "${ucloud_anti_ddos_instance.foo.id}"
    ip = "${ucloud_anti_ddos_ip.foo.ip}"
    real_server_type = "IP"
	real_servers {
      address   = "127.0.0.1"
    }
    real_servers {
      address   = "127.0.0.2"
    }
	toa_id = 100
}
`
