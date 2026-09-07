package uads_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	sdkuads "github.com/ucloud/ucloud-sdk-go/services/uads"
)

func TestAccUCloudAntiDDoS_basic(t *testing.T) {
	var uadsServiceInfo sdkuads.ServiceInfo

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_anti_ddos_instance.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAntiDDoSInstanceDestroy,

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
