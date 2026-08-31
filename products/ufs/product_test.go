package ufs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func TestRegistration(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New())); err != nil {
		t.Fatalf("register UFS product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate provider with UFS product: %v", err)
	}

	for _, name := range []string{"ucloud_ufs_volume", "ucloud_ufs_volume_mount_point"} {
		if provider.ResourcesMap[name] == nil {
			t.Fatalf("%s is not registered", name)
		}
	}
	if provider.DataSourcesMap["ucloud_ufs_volumes"] == nil {
		t.Fatal("ucloud_ufs_volumes is not registered")
	}
}

func TestVolumeSchemaCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_ufs_volume"]
	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
	}{
		"size":          {typeValue: schema.TypeInt, required: true},
		"storage_type":  {typeValue: schema.TypeString, required: true},
		"protocol_type": {typeValue: schema.TypeString, required: true},
		"name":          {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"charge_type":   {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"duration":      {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"tag":           {typeValue: schema.TypeString, optional: true, forceNew: true},
		"remark":        {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"create_time":   {typeValue: schema.TypeString, computed: true},
		"expire_time":   {typeValue: schema.TypeString, computed: true},
	}
	if len(resource.Schema) != len(wantFields) {
		t.Fatalf("volume schema has %d fields, want %d", len(resource.Schema), len(wantFields))
	}
	for name, want := range wantFields {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("volume schema is missing %q", name)
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Optional != want.optional || field.Computed != want.computed || field.ForceNew != want.forceNew {
			t.Errorf("volume field %q flags = type=%v required=%t optional=%t computed=%t force_new=%t", name, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew)
		}
	}

	if resource.Schema["tag"].Default != defaultTag {
		t.Errorf("tag default = %#v, want %q", resource.Schema["tag"].Default, defaultTag)
	}
	for value, want := range map[string]string{"": defaultTag, "team-a": "team-a"} {
		if got := resource.Schema["tag"].StateFunc(value); got != want {
			t.Errorf("tag StateFunc(%q) = %q, want %q", value, got, want)
		}
	}

	validate := resource.Schema["size"].ValidateFunc
	for value, wantErr := range map[int]bool{100: false, 500: false, 550: true, 100001: true} {
		_, errors := validate(value, "size")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("size validation errors for %d = %v, wantErr = %t", value, errors, wantErr)
		}
	}
}

func TestVolumeMountPointSchemaCompatibility(t *testing.T) {
	resource := New().Registration().Resources["ucloud_ufs_volume_mount_point"]
	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		computed  bool
		forceNew  bool
	}{
		"volume_id":      {typeValue: schema.TypeString, required: true, forceNew: true},
		"name":           {typeValue: schema.TypeString, required: true, forceNew: true},
		"vpc_id":         {typeValue: schema.TypeString, required: true, forceNew: true},
		"subnet_id":      {typeValue: schema.TypeString, required: true, forceNew: true},
		"mount_point_ip": {typeValue: schema.TypeString, computed: true},
		"create_time":    {typeValue: schema.TypeString, computed: true},
	}
	if len(resource.Schema) != len(wantFields) {
		t.Fatalf("mount point schema has %d fields, want %d", len(resource.Schema), len(wantFields))
	}
	for name, want := range wantFields {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("mount point schema is missing %q", name)
		}
		if field.Type != want.typeValue || field.Required != want.required || field.Computed != want.computed || field.ForceNew != want.forceNew {
			t.Errorf("mount point field %q flags = type=%v required=%t computed=%t force_new=%t", name, field.Type, field.Required, field.Computed, field.ForceNew)
		}
	}
}

func TestVolumesDataSourceSchemaCompatibility(t *testing.T) {
	dataSource := New().Registration().DataSources["ucloud_ufs_volumes"]
	for _, name := range []string{"ids", "name_regex", "output_file", "total_count", "ufs_volumes"} {
		if dataSource.Schema[name] == nil {
			t.Fatalf("data source schema is missing %q", name)
		}
	}
	if dataSource.Schema["ids"].Type != schema.TypeSet || !dataSource.Schema["ids"].Optional || !dataSource.Schema["ids"].Computed {
		t.Fatal("ids must remain an optional computed TypeSet")
	}
	if dataSource.Schema["ufs_volumes"].Type != schema.TypeList || !dataSource.Schema["ufs_volumes"].Computed {
		t.Fatal("ufs_volumes must remain a computed TypeList")
	}
	nested, ok := dataSource.Schema["ufs_volumes"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("ufs_volumes element type = %T, want *schema.Resource", dataSource.Schema["ufs_volumes"].Elem)
	}
	for _, name := range []string{"id", "name", "tag", "remark", "size", "storage_type", "protocol_type", "create_time", "expire_time"} {
		if nested.Schema[name] == nil || !nested.Schema[name].Computed {
			t.Fatalf("ufs_volumes nested field %q must be computed", name)
		}
	}
}

func TestChargeTypeConversionCompatibility(t *testing.T) {
	for value, want := range map[string]string{
		"year":    "Year",
		"month":   "Month",
		"dynamic": "Dynamic",
	} {
		if got := chargeTypeToAPI(value); got != want {
			t.Errorf("chargeTypeToAPI(%q) = %q, want %q", value, got, want)
		}
	}
}
