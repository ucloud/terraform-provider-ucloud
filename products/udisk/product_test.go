package udisk

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("disk", "disks"))); err != nil {
		t.Fatalf("register udisk: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate udisk provider: %v", err)
	}

	for _, name := range []string{
		"ucloud_disk",
		"ucloud_disk_attachment",
		"ucloud_disk_snapshot",
	} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_disks", "ucloud_disk_snapshots"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestRegistrationUsesStableSchemaFields(t *testing.T) {
	disk := resourceUCloudDisk()
	for _, field := range []string{"availability_zone", "disk_size", "disk_type", "snapshot_id", "snapshot_service", "charge_type", "tag"} {
		if disk.Schema[field] == nil {
			t.Errorf("disk schema is missing field %q", field)
		}
	}

	attachment := resourceUCloudDiskAttachment()
	for _, field := range []string{"availability_zone", "instance_id", "disk_id", "stop_instance_before_detaching", "device_name"} {
		if attachment.Schema[field] == nil {
			t.Errorf("attachment schema is missing field %q", field)
		}
	}

	snapshot := resourceUCloudDiskSnapshot()
	for _, field := range []string{"availability_zone", "disk_id", "name", "comment", "disk_type", "size", "source_disk_name", "is_disk_available", "create_time", "status"} {
		if snapshot.Schema[field] == nil {
			t.Errorf("snapshot schema is missing field %q", field)
		}
	}
}

type runtimeStub struct {
	name  string
	calls int
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
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
	if stub.name != Name {
		t.Fatalf("product client name = %q, want %q", stub.name, Name)
	}
	if stub.calls != 1 {
		t.Fatalf("product client calls = %d, want 1", stub.calls)
	}
	if client == nil || client.udiskconn == nil || client.uhostconn == nil {
		t.Fatal("product client did not initialize UDisk and UHost SDK clients")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}
