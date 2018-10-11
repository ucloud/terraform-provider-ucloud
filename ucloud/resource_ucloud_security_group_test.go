package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudSecurityGroup_basic(t *testing.T) {
	var sgSet unet.FirewallDataSet

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_security_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecurityGroupDestroy,

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccSecurityGroupConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("ucloud_security_group.foo", &sgSet),
					testAccCheckSecurityGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "name", "testAcc5"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.99862869.port_range", "80"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.99862869.protocol", "TCP"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.99862869.cidr_block", "192.168.0.0/16"),
				),
			},

			resource.TestStep{
				Config: testAccSecurityGroupConfigTwo,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("ucloud_security_group.foo", &sgSet),
					testAccCheckSecurityGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "name", "testAccTwo"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2859557110.port_range", "20-80"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2859557110.protocol", "TCP"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2859557110.cidr_block", "0.0.0.0/0"),
				),
			},
		},
	})

}

func testAccCheckSecurityGroupExists(n string, sgSet *unet.FirewallDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("security group id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeFirewallById(rs.Primary.ID)

		log.Printf("[INFO] security group id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*sgSet = *ptr
		return nil
	}
}

func testAccCheckSecurityGroupAttributes(sgSet *unet.FirewallDataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sgSet.FWId == "" {
			return fmt.Errorf("security group id is empty")
		}
		return nil
	}
}

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_security_group" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeFirewallById(rs.Primary.ID)

		// Verify the error is what we want
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.FWId != "" {
			return fmt.Errorf("security group still exist")
		}
	}

	return nil
}

const testAccSecurityGroupConfig = `
resource "ucloud_security_group" "foo" {
	name = "testAcc5"
	rules {
		port_range = "80"
		protocol   = "TCP"
		cidr_block = "192.168.0.0/16"
	}
}
`
const testAccSecurityGroupConfigTwo = `
resource "ucloud_security_group" "foo" {
	name = "testAccTwo"
	rules {
		port_range = "20-80"
		protocol   = "TCP"
		cidr_block = "0.0.0.0/0"
	}
}
`

func Test_resourceucloudSecurityGroupRuleHash(t *testing.T) {
	m := map[string]interface{}{
		"port_range": "80",
		"protocol":   "TCP",
		"cidr_block": "192.168.0.0/16",
		"priority":   "HIGH",
		"policy":     "ACCEPT",
	}
	want := 99862869
	got := resourceucloudSecurityGroupRuleHash(m)
	if want != got {
		t.Errorf("resourceucloudSecurityGroupRuleHash() = %v, want %v", got, want)
	}
}
