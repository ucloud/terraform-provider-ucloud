package ucloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"log"
	"testing"
)

func TestAccUCloudUK8SNode_basic(t *testing.T) {
	var uk8sNodeSet uk8s.NodeInfoV2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_uk8s_node.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUK8SNodeDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccUK8SNodeConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckUK8SNodeExists("ucloud_uk8s_node.foo", &uk8sNodeSet),
					testAccCheckUK8SNodeAttributes(&uk8sNodeSet),
					resource.TestCheckResourceAttr("ucloud_uk8s_node.foo", "ip_set.#", "1"),
					resource.TestCheckResourceAttr("ucloud_uk8s_node.foo", "status", "Ready"),
				),
			},
		},
	})
}

func testAccCheckUK8SNodeExists(n string, uk8sNodeSet *uk8s.NodeInfoV2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("uk8s cluster id is empty")
		}

		client := testAccProvider.Meta().(*UCloudClient)
		ptr, err := client.describeUK8SClusterNodeByResourceId(rs.Primary.Attributes["cluster_id"], rs.Primary.ID)

		log.Printf("[INFO] uk8s cluster id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		*uk8sNodeSet = *ptr
		return nil
	}
}

func testAccCheckUK8SNodeAttributes(uk8sNodeSet *uk8s.NodeInfoV2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if uk8sNodeSet.NodeId == "" {
			return fmt.Errorf("uk8s node id is empty")
		}
		return nil
	}
}

func testAccCheckUK8SNodeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ucloud_uk8s_node" {
			continue
		}

		client := testAccProvider.Meta().(*UCloudClient)
		d, err := client.describeUK8SClusterNodeByResourceId(rs.Primary.Attributes["cluster_id"], rs.Primary.ID)
		if err != nil {
			if isNotFoundError(err) {
				continue
			}
			return err
		}

		if d.NodeId != "" {
			return fmt.Errorf("uk8s node still exist")
		}
	}
	return nil
}

const testAccUK8SNodeConfig = `
resource "ucloud_vpc" "foo" {
  name        = "tf-acc-uk8s-node"
  tag         = "tf-acc"
  cidr_blocks = ["192.168.0.0/16"]
}

resource "ucloud_subnet" "foo" {
  name       = "tf-acc-uk8s-node"
  tag        = "tf-acc"
  cidr_block = "192.168.1.0/24"
  vpc_id     = "${ucloud_vpc.foo.id}"
}

data "ucloud_zones" "default" {
}

resource "ucloud_uk8s_cluster" "foo" {
  vpc_id       = "${ucloud_vpc.foo.id}"
  subnet_id    = "${ucloud_subnet.foo.id}"
  name         = "tf-acc-uk8s-node-basic"
  service_cidr = "172.16.0.0/16"
  password     = "ucloud_2021"
  charge_type  = "dynamic"

  master {
    availability_zones = [
      "${data.ucloud_zones.default.zones.0.id}",
      "${data.ucloud_zones.default.zones.0.id}",
      "${data.ucloud_zones.default.zones.0.id}",
    ]
    instance_type = "n-basic-2"
  }
}

resource "ucloud_uk8s_node" "foo" {
  cluster_id    = "${ucloud_uk8s_cluster.foo.id}"
  subnet_id     = "${ucloud_subnet.foo.id}"
  password      = "ucloud_2021"
  instance_type = "n-basic-2"
  charge_type   = "dynamic"
  availability_zone = "${data.ucloud_zones.default.zones.0.id}"
  
  count = 2
}
`
