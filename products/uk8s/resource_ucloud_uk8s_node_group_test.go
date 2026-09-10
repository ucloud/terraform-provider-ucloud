package uk8s_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func TestAccUCloudUK8SNodeGroup_basic(t *testing.T) {
	const name = "ucloud_uk8s_node_group.workers"
	var originalID string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUK8SNodeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUK8SNodeGroupConfig("workers", "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "workers"),
					resource.TestCheckResourceAttr(name, "labels.environment", "test"),
					resource.TestCheckResourceAttr(name, "taint.#", "1"),
					func(state *terraform.State) error {
						originalID = state.RootModule().Resources[name].Primary.ID
						if originalID == "" {
							return fmt.Errorf("node group has no ID")
						}
						return nil
					},
				),
			},
			{
				ResourceName: name, ImportState: true, ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					item := state.RootModule().Resources[name].Primary
					return item.Attributes["cluster_id"] + "/" + item.ID, nil
				},
			},
			{
				Config: testAccUK8SNodeGroupConfig("workers-updated", "production"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "workers-updated"),
					resource.TestCheckResourceAttr(name, "labels.environment", "production"),
					func(state *terraform.State) error {
						if id := state.RootModule().Resources[name].Primary.ID; id != originalID {
							return fmt.Errorf("node group %q was replaced by %q", originalID, id)
						}
						return nil
					},
				),
			},
			{
				Config: testAccUK8SClusterConfig,
				Check: func(state *terraform.State) error {
					client, err := testAccUK8SClient()
					if err != nil {
						return err
					}
					req := client.NewListUK8SNodeGroupRequest()
					req.ClusterId = ucloud.String(state.RootModule().Resources["ucloud_uk8s_cluster.foo"].Primary.ID)
					resp, err := client.ListUK8SNodeGroup(req)
					if err != nil {
						return fmt.Errorf("check deleted node group %q: %w", originalID, err)
					}
					for _, group := range resp.NodeGroupList {
						if group.NodeGroupId == originalID {
							return fmt.Errorf("node group %q still exists after deletion", originalID)
						}
					}
					return nil
				},
			},
		},
	})
}

func testAccCheckUK8SNodeGroupDestroy(state *terraform.State) error {
	client, err := testAccUK8SClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_uk8s_node_group" {
			continue
		}
		clusterID := item.Primary.Attributes["cluster_id"]
		_, found, err := describeAccUK8SClusterByID(client, clusterID)
		if err != nil {
			return fmt.Errorf("check deleted node group %q: %w", item.Primary.ID, err)
		}
		if !found {
			continue
		}
		req := client.NewListUK8SNodeGroupRequest()
		req.ClusterId = ucloud.String(clusterID)
		resp, err := client.ListUK8SNodeGroup(req)
		if err != nil {
			return fmt.Errorf("check deleted node group %q: %w", item.Primary.ID, err)
		}
		for _, group := range resp.NodeGroupList {
			if group.NodeGroupId == item.Primary.ID {
				return fmt.Errorf("node group %q still exists", item.Primary.ID)
			}
		}
	}
	return testAccCheckUK8SClusterDestroy(state)
}

func testAccUK8SNodeGroupConfig(name, environment string) string {
	return testAccUK8SClusterConfig + fmt.Sprintf(`
resource "ucloud_uk8s_node_group" "workers" {
  cluster_id        = ucloud_uk8s_cluster.foo.id
  name              = %q
  availability_zone = data.ucloud_zones.default.zones[0].id
  subnet_id         = ucloud_subnet.foo.id
  instance_type     = "n-basic-2"
  charge_type       = "dynamic"
  labels = { environment = %q }
  taint {
    key    = "dedicated"
    value  = "test"
    effect = "NoSchedule"
  }
}
`, name, environment)
}
