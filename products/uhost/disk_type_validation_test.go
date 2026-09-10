package uhost

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestInstanceDiskTypeValidators(t *testing.T) {
	instance := resourceUCloudInstance()

	dataDisks, ok := instance.Schema["data_disks"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("data_disks element type = %T, want *schema.Resource", instance.Schema["data_disks"].Elem)
	}

	tests := []struct {
		name         string
		field        *schema.Schema
		value        string
		wantValidate bool
	}{
		{name: "data disk type accepts cloud_essd", field: dataDisks.Schema["type"], value: "cloud_essd"},
		{name: "data disk type accepts cloud_rssd", field: dataDisks.Schema["type"], value: "cloud_rssd"},
		{name: "data disk type accepts cloud_ssd", field: dataDisks.Schema["type"], value: "cloud_ssd"},
		{name: "data disk type accepts cloud_normal", field: dataDisks.Schema["type"], value: "cloud_normal"},
		{
			name:         "data disk type rejects local disk",
			field:        dataDisks.Schema["type"],
			value:        "local_ssd",
			wantValidate: true,
		},
		{
			name:         "boot disk type rejects cloud_essd",
			field:        instance.Schema["boot_disk_type"],
			value:        "cloud_essd",
			wantValidate: true,
		},
		{name: "boot disk type accepts cloud_rssd", field: instance.Schema["boot_disk_type"], value: "cloud_rssd"},
		{
			name:         "legacy data disk type rejects cloud_essd",
			field:        instance.Schema["data_disk_type"],
			value:        "cloud_essd",
			wantValidate: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			validate := test.field.ValidateFunc
			if validate == nil {
				t.Fatal("field has no ValidateFunc")
			}
			_, errs := validate(test.value, "type")
			if (len(errs) > 0) != test.wantValidate {
				t.Fatalf("validate(%q) errors = %v, wantErr %v", test.value, errs, test.wantValidate)
			}
		})
	}
}
