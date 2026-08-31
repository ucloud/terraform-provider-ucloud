package unet_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
)

func TestAccUCloudEIPAssociation_basic(t *testing.T) {
	var eip unet.UnetEIPSet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		IDRefreshName: "ucloud_eip_association.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPAssociationDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccEIPAssociationConfig,

				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists("ucloud_eip.foo", &eip),
					testAccCheckEIPAssociationExists("ucloud_eip_association.foo"),
				),
			},
		},
	})
}

func testAccCheckEIPAssociationExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("eip association id is empty")
		}

		eipID := item.Primary.Attributes["eip_id"]
		resourceID := item.Primary.Attributes["resource_id"]
		client, err := testAccUNetClient()
		if err != nil {
			return err
		}

		return resource.Retry(3*time.Minute, func() *resource.RetryError {
			binding, found, err := describeAccEIPResourceByID(client, eipID, resourceID)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			if !found || (binding.ResourceID != resourceID && binding.ResourceId != resourceID) {
				return resource.NonRetryableError(fmt.Errorf("eip association not found"))
			}
			return nil
		})
	}
}

const testAccEIPAssociationConfig = `
data "ucloud_zones" "default" {}

data "ucloud_images" "default" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	name_regex        = "^CentOS 7.[1-2] 64"
	image_type        = "base"
}

resource "ucloud_eip" "foo" {
	name          = "tf-acc-eip-association-eip"
	tag           = "tf-acc"
	internet_type = "bgp"
	bandwidth     = 1
	duration      = 1
}

data "ucloud_security_groups" "default" {
	type = "recommend_web"
}

resource "ucloud_instance" "foo" {
	availability_zone = "${data.ucloud_zones.default.zones.0.id}"
	image_id          = "${data.ucloud_images.default.images.0.id}"
	security_group    = "${data.ucloud_security_groups.default.security_groups.0.id}"
	instance_type     = "n-basic-1"
	root_password     = "wA1234567"
	name              = "tf-acc-eip-association-instance"
	tag               = "tf-acc"
}

resource "ucloud_eip_association" "foo" {
	eip_id        = "${ucloud_eip.foo.id}"
	resource_type = "instance"
	resource_id   = "${ucloud_instance.foo.id}"
}
`

func TestEIPAssociationAcceptanceConfigUsesRegisteredTypes(t *testing.T) {
	file, err := hcl.Parse(testAccEIPAssociationConfig)
	if err != nil {
		t.Fatalf("parse acceptance configuration: %v", err)
	}
	objects, ok := file.Node.(*ast.ObjectList)
	if !ok {
		t.Fatalf("acceptance configuration root has type %T, want *ast.ObjectList", file.Node)
	}

	registrations := []struct {
		kind  string
		types map[string]*schema.Resource
	}{
		{kind: "resource", types: testAccHarness.Provider.ResourcesMap},
		{kind: "data", types: testAccHarness.Provider.DataSourcesMap},
	}
	for _, registration := range registrations {
		items := objects.Filter(registration.kind).Items
		if len(items) == 0 {
			t.Fatalf("acceptance configuration contains no %s declarations", registration.kind)
		}
		for _, item := range items {
			if len(item.Keys) < 2 {
				t.Fatalf("%s declaration has %d labels, want at least 2", registration.kind, len(item.Keys))
			}
			typeName, ok := item.Keys[0].Token.Value().(string)
			if !ok {
				t.Fatalf("%s type label has value type %T, want string", registration.kind, item.Keys[0].Token.Value())
			}
			if _, registered := registration.types[typeName]; !registered {
				t.Errorf("acceptance configuration references unregistered %s type %q", registration.kind, typeName)
			}
		}
	}
}
