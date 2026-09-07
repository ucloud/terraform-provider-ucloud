package unet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	sdkunet "github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("eip", "eips", "security_group", "security_groups"))); err != nil {
		t.Fatalf("register unet: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate unet provider: %v", err)
	}

	for _, name := range []string{
		"ucloud_eip",
		"ucloud_eip_association",
		"ucloud_security_group",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_eips", "ucloud_security_groups"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestResourceSchemasKeepCompatibilityFields(t *testing.T) {
	eip := resourceUCloudEIP()
	for _, field := range []string{
		"bandwidth", "internet_type", "charge_type", "charge_mode", "share_bandwidth_package_id",
		"duration", "name", "remark", "tag", "status", "public_ip", "ip_set", "resource", "create_time", "expire_time",
	} {
		if eip.Schema[field] == nil {
			t.Errorf("eip schema is missing field %q", field)
		}
	}
	if !eip.Schema["internet_type"].Required || !eip.Schema["internet_type"].ForceNew {
		t.Error("eip internet_type must remain required and ForceNew")
	}
	if !eip.Schema["charge_type"].Optional || !eip.Schema["charge_type"].Computed || !eip.Schema["charge_type"].ForceNew {
		t.Error("eip charge_type optional/computed/ForceNew flags changed")
	}
	if got := eip.Schema["tag"].Default; got != defaultTag {
		t.Errorf("eip tag default = %#v, want %q", got, defaultTag)
	}

	association := resourceUCloudEIPAssociation()
	if association.SchemaVersion != 1 || association.MigrateState == nil {
		t.Error("eip association state migration is missing")
	}
	for _, field := range []string{"eip_id", "resource_type", "resource_id"} {
		if association.Schema[field] == nil {
			t.Errorf("eip association schema is missing field %q", field)
		}
	}
	if !association.Schema["eip_id"].Required || !association.Schema["eip_id"].ForceNew {
		t.Error("eip association eip_id must remain required and ForceNew")
	}

	securityGroup := resourceUCloudSecurityGroup()
	if securityGroup.Create == nil || securityGroup.Read == nil || securityGroup.Update == nil || securityGroup.Delete == nil {
		t.Error("security group CRUD callbacks are incomplete")
	}
	if securityGroup.Importer == nil || securityGroup.CustomizeDiff == nil {
		t.Error("security group importer or CustomizeDiff is missing")
	}
	for _, field := range []string{"name", "rules", "tag", "remark", "create_time"} {
		if securityGroup.Schema[field] == nil {
			t.Errorf("security group schema is missing field %q", field)
		}
	}
	if !securityGroup.Schema["rules"].Required || securityGroup.Schema["rules"].Type != schema.TypeSet {
		t.Error("security group rules must remain a required TypeSet")
	}
}

func TestDataSourceSchemasKeepCompatibilityFields(t *testing.T) {
	eips := dataSourceUCloudEips()
	if eips.SchemaVersion != 0 || eips.MigrateState != nil {
		t.Error("eips data source must not enable the legacy orphan state migration")
	}
	for _, field := range []string{"ids", "name_regex", "output_file", "total_count", "eips"} {
		if eips.Schema[field] == nil {
			t.Errorf("eips data source schema is missing field %q", field)
		}
	}

	securityGroups := dataSourceUCloudSecurityGroups()
	for _, field := range []string{"ids", "name_regex", "type", "output_file", "total_count", "security_groups"} {
		if securityGroups.Schema[field] == nil {
			t.Errorf("security groups data source schema is missing field %q", field)
		}
	}
}

func TestCompatibilityStateMigrations(t *testing.T) {
	association := &terraform.InstanceState{
		ID: "eip#eip-abcd:instance#uhost-abcd",
		Attributes: map[string]string{
			"id": "eip#eip-abcd:instance#uhost-abcd",
		},
	}
	if _, err := resourceUCloudEIPAssociationMigrateState(0, association, nil); err != nil {
		t.Fatalf("migrate eip association: %v", err)
	}
	if association.ID != "eip-abcd:uhost-abcd" || association.Attributes["id"] != association.ID {
		t.Fatalf("migrated association = %#v, want eip-abcd:uhost-abcd", association)
	}

	eips := &terraform.InstanceState{
		ID: "eips",
		Attributes: map[string]string{
			"eips.0.charge_type": "Month",
			"eips.1.charge_mode": "ShareBandwidth",
		},
	}
	if _, err := dataSourceUCloudEipsMigrateState(0, eips, nil); err != nil {
		t.Fatalf("migrate eips data source: %v", err)
	}
	if eips.Attributes["eips.0.charge_type"] != "month" || eips.Attributes["eips.1.charge_mode"] != "share_bandwidth" {
		t.Fatalf("migrated eips state = %#v", eips.Attributes)
	}
}

type runtimeStub struct {
	name  string
	calls int
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	if name != Name {
		return nil, fmt.Errorf("unexpected product %q", name)
	}
	stub.name = name
	stub.calls++
	config := ucloud.NewConfig()
	return constructor(&config, &auth.Credential{}, nil), nil
}

func TestClientFromMetaUsesProductRuntime(t *testing.T) {
	stub := &runtimeStub{}
	client, err := clientFromMeta(stub)
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}
	if stub.name != Name || stub.calls != 1 {
		t.Fatalf("runtime call = (%q, %d), want (%q, 1)", stub.name, stub.calls, Name)
	}
	if _, ok := interface{}(client).(*sdkunet.UNetClient); !ok || client == nil {
		t.Fatal("product client did not initialize the UNet SDK client")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}
