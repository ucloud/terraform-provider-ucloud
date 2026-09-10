package uk8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

func masterTestConfig(master map[string]interface{}) map[string]interface{} {
	master["availability_zones"] = []interface{}{"sg-02", "sg-02", "sg-02"}
	return map[string]interface{}{
		"name": "test", "service_cidr": "192.168.0.0/16",
		"vpc_id": "uvnet-test", "subnet_id": "subnet-test", "password": "TestPassword123",
		"image_id": "uimage-default", "master": []interface{}{master},
	}
}

func TestMasterCreateRequest(t *testing.T) {
	for _, test := range []struct {
		name    string
		cniMode string
		master  map[string]interface{}
		want    map[string]string
		absent  []string
	}{
		{
			name:    "console specification",
			cniMode: "VPC",
			master: map[string]interface{}{
				"machine_type": "O", "cpu": 2, "memory": 4096, "uhost_family": "o2i",
				"min_cpu_platform": "Intel/EmeraldRapids", "image_id": "uimage-1rgr4ndomwqa",
				"boot_disk_type": "cloud_rssd", "boot_disk_size": 40,
				"data_disk_type": "cloud_rssd", "data_disk_size": 20,
			},
			want: map[string]string{
				"CNIMode":           "VPC",
				"MasterMachineType": "O", "MasterCPU": "2", "MasterMem": "4096",
				"MasterUHostFamily": "o2i", "MasterMinimalCpuPlatform": "Intel/EmeraldRapids",
				"MasterImageId": "uimage-1rgr4ndomwqa", "MasterBootDiskType": "CLOUD_RSSD",
				"MasterBootDiskSize": "40", "MasterDataDiskType": "CLOUD_RSSD", "MasterDataDiskSize": "20",
			},
		},
		{
			name:   "legacy specification and image fallback",
			master: map[string]interface{}{"instance_type": "n-basic-2"},
			want:   map[string]string{"MasterMachineType": "N", "MasterCPU": "2", "MasterMem": "4096", "MasterMinimalCpuPlatform": "Intel/Auto"},
			absent: []string{"CNIMode", "MasterUHostFamily", "MasterImageId", "MasterBootDiskSize"},
		},
		{
			name:    "Calico network",
			cniMode: "Calico",
			master:  map[string]interface{}{"machine_type": "N", "cpu": 2, "memory": 4096},
			want:    map[string]string{"CNIMode": "Calico"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := resourceUCloudUK8SCluster()
			config := masterTestConfig(test.master)
			if test.cniMode != "" {
				config["cni_mode"] = test.cniMode
			}
			if _, errs := r.Validate(terraform.NewResourceConfigRaw(config)); len(errs) > 0 {
				t.Fatalf("validate: %v", errs)
			}
			d := schema.TestResourceDataRaw(t, r.Schema, config)
			cfg := ucloud.NewConfig()
			cfg.Region = "sg"
			cfg.ProjectId = "org-test"
			client := sdkuk8s.NewClient(&cfg, &auth.Credential{})
			req, err := buildUK8SClusterCreateRequest(client, d)
			if err != nil {
				t.Fatal(err)
			}
			if req.GetRetryable() {
				t.Fatal("create request must not be automatically retried")
			}
			form, err := request.EncodeForm(req)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range map[string]string{
				"Action": "CreateUK8SClusterV2", "Region": "sg", "ProjectId": "org-test",
				"ImageId": "uimage-default", "Master.0.Zone": "sg-02", "Master.1.Zone": "sg-02", "Master.2.Zone": "sg-02",
			} {
				test.want[k] = v
			}
			for key, want := range test.want {
				if form[key] != want {
					t.Errorf("%s = %q, want %q", key, form[key], want)
				}
			}
			for _, key := range append(test.absent, "MasterMinmalCpuPlatform", "InstanceType") {
				if _, ok := form[key]; ok {
					t.Errorf("unexpected parameter %s", key)
				}
			}
		})
	}
}

func TestMasterConfigurationValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		master map[string]interface{}
	}{
		{name: "conflicting specification", master: map[string]interface{}{"instance_type": "n-basic-2", "machine_type": "N", "cpu": 2, "memory": 4096}},
		{name: "missing specification", master: map[string]interface{}{}},
		{name: "partial specification", master: map[string]interface{}{"machine_type": "O", "cpu": 2}},
		{name: "memory unit", master: map[string]interface{}{"machine_type": "O", "cpu": 2, "memory": 4}},
		{name: "memory alignment", master: map[string]interface{}{"machine_type": "O", "cpu": 2, "memory": 4097}},
		{name: "boot disk size", master: map[string]interface{}{"machine_type": "N", "cpu": 2, "memory": 4096, "boot_disk_size": 20}},
		{name: "family machine mismatch", master: map[string]interface{}{"machine_type": "N", "cpu": 2, "memory": 4096, "uhost_family": "o2i"}},
		{name: "family platform mismatch", master: map[string]interface{}{"machine_type": "O", "cpu": 2, "memory": 4096, "uhost_family": "o2i", "boot_disk_type": "cloud_rssd", "min_cpu_platform": "Amd/Auto"}},
		{name: "outstanding disk mismatch", master: map[string]interface{}{"machine_type": "O", "cpu": 2, "memory": 4096, "boot_disk_type": "cloud_ssd"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := resourceUCloudUK8SCluster()
			config := terraform.NewResourceConfigRaw(masterTestConfig(test.master))
			_, errs := r.Validate(config)
			if len(errs) > 0 {
				return
			}
			if _, err := r.Diff(nil, config, nil); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}

func TestClusterCNIModePlan(t *testing.T) {
	for _, test := range []struct {
		name        string
		mode        string
		wantReplace bool
		wantInvalid bool
	}{
		{name: "preserve API selected mode"},
		{name: "configure existing mode", mode: "VPC"},
		{name: "change mode", mode: "Calico", wantReplace: true},
		{name: "invalid mode", mode: "unknown", wantInvalid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := resourceUCloudUK8SCluster()
			config := masterTestConfig(map[string]interface{}{"machine_type": "N", "cpu": 2, "memory": 4096})
			d := schema.TestResourceDataRaw(t, r.Schema, config)
			d.SetId("uk8s-test")
			if err := d.Set("cni_mode", "VPC"); err != nil {
				t.Fatal(err)
			}
			if test.mode != "" {
				config["cni_mode"] = test.mode
			}
			raw := terraform.NewResourceConfigRaw(config)
			_, errs := r.Validate(raw)
			if (len(errs) > 0) != test.wantInvalid {
				t.Fatalf("validation errors = %v, wantInvalid = %v", errs, test.wantInvalid)
			}
			if test.wantInvalid {
				return
			}
			diff, err := r.Diff(d.State(), raw, nil)
			if err != nil {
				t.Fatal(err)
			}
			if test.wantReplace {
				if diff == nil || !diff.RequiresNew() {
					t.Fatal("changing CNI mode must require replacement")
				}
			} else if diff != nil && !diff.Empty() {
				t.Fatalf("unchanged CNI mode produced a diff: %#v", diff)
			}
		})
	}
}
