package ucloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudSecurityGroup_basic(t *testing.T) {
	rInt := acctest.RandInt()
	var sgSet unet.FirewallDataSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_security_group.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSecurityGroupDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfig(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("ucloud_security_group.foo", &sgSet),
					testAccCheckSecurityGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "name", fmt.Sprintf("tf-acc-security-group-%d", rInt)),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "tag", "tf-acc"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2629295509.port_range", "80"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2629295509.protocol", "tcp"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.2629295509.cidr_block", "192.168.0.0/16"),
				),
			},

			{
				Config: testAccSecurityGroupConfigTwo(rInt),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("ucloud_security_group.foo", &sgSet),
					testAccCheckSecurityGroupAttributes(&sgSet),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "name", fmt.Sprintf("tf-acc-security-group-%d-two", rInt)),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "tag", defaultTag),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.3266055183.port_range", "20-80"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.3266055183.protocol", "tcp"),
					resource.TestCheckResourceAttr("ucloud_security_group.foo", "rules.3266055183.cidr_block", "0.0.0.0/0"),
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

func testAccSecurityGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_security_group" "foo" {
	name = "tf-acc-security-group-%d"
	tag  = "tf-acc"
	rules {
		port_range = "80"
		protocol   = "tcp"
		cidr_block = "192.168.0.0/16"
		policy     = "accept"
		priority   = "high"
	}
}`, rInt)
}

func testAccSecurityGroupConfigTwo(rInt int) string {
	return fmt.Sprintf(`
resource "ucloud_security_group" "foo" {
	name = "tf-acc-security-group-%d-two"
	tag  = ""
	rules {
		port_range = "20-80"
		protocol   = "tcp"
		cidr_block = "0.0.0.0/0"
		policy     = "accept"
		priority   = "high"
	}
}`, rInt)
}

func Test_resourceucloudSecurityGroupRuleHash(t *testing.T) {
	m := map[string]interface{}{
		"port_range": "80",
		"protocol":   "tcp",
		"cidr_block": "192.168.0.0/16",
		"policy":     "accept",
		"priority":   "high",
	}

	want := 2629295509
	got := resourceucloudSecurityGroupRuleHash(m)
	if want != got {
		t.Errorf("resourceucloudSecurityGroupRuleHash() = %v, want %v", got, want)
	}
}
