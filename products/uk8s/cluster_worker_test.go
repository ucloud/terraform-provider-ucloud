package uk8s

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

func workerTestConfig(workers ...interface{}) map[string]interface{} {
	config := masterTestConfig(map[string]interface{}{"machine_type": "N", "cpu": 2, "memory": 4096})
	if len(workers) > 0 {
		config["worker"] = workers
	}
	return config
}

func TestClusterWorkerRequest(t *testing.T) {
	for _, test := range []struct {
		name    string
		workers []interface{}
		want    map[string]string
		absent  []string
	}{
		{name: "omitted"},
		{
			name: "console and second zone",
			workers: []interface{}{
				map[string]interface{}{
					"availability_zone": "sg-02", "machine_type": "O", "cpu": 2, "memory": 4096,
					"count": 1, "gpu": 0, "uhost_family": "o2i", "min_cpu_platform": "Intel/EmeraldRapids",
					"boot_disk_type": "cloud_rssd", "boot_disk_size": 40,
					"data_disk_type": "cloud_rssd", "data_disk_size": 20, "net_capability": "Ultra",
					"max_pods": 110, "image_id": "uimage-1rgr4ndomwqa", "security_group_id": "291852",
					"labels": map[string]interface{}{"role": "worker", "env": "test"},
					"taint":  []interface{}{map[string]interface{}{"key": "dedicated", "value": "test", "effect": "NoSchedule"}},
				},
				map[string]interface{}{"availability_zone": "sg-01", "machine_type": "N", "cpu": 4, "memory": 8192, "count": 2},
			},
			want: map[string]string{
				"Nodes.0.Zone": "sg-02", "Nodes.0.MachineType": "O", "Nodes.0.CPU": "2", "Nodes.0.Mem": "4096",
				"Nodes.0.Count": "1", "Nodes.0.GPU": "0", "Nodes.0.UHostFamily": "o2i",
				"Nodes.0.MinimalCpuPlatform": "Intel/EmeraldRapids", "Nodes.0.BootDiskType": "CLOUD_RSSD",
				"Nodes.0.BootDiskSize": "40", "Nodes.0.DataDiskType": "CLOUD_RSSD", "Nodes.0.DataDiskSize": "20",
				"Nodes.0.NetCapability": "Ultra", "Nodes.0.MaxPods": "110", "Nodes.0.ImageId": "uimage-1rgr4ndomwqa",
				"Nodes.0.SecurityGroupId": "291852", "Nodes.0.Labels": "env=test,role=worker", "Nodes.0.Taints": "dedicated=test:NoSchedule",
				"Nodes.1.Zone": "sg-01", "Nodes.1.MachineType": "N", "Nodes.1.CPU": "4", "Nodes.1.Mem": "8192", "Nodes.1.Count": "2",
				"Nodes.1.BootDiskSize": "40", "Nodes.1.BootDiskType": "CLOUD_SSD", "Nodes.1.MinimalCpuPlatform": "Intel/Auto",
				"ImageId": "uimage-default",
			},
			absent: []string{"Nodes.0.BootDiskSIze", "Nodes.0.MinmalCpuPlatform", "Nodes.1.ImageId", "Nodes.1.UHostFamily", "Nodes.1.DataDiskSize", "Nodes.1.Labels", "Nodes.1.Taints", "Nodes.2.Zone"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := resourceUCloudUK8SCluster()
			config := workerTestConfig(test.workers...)
			if _, errs := r.Validate(terraform.NewResourceConfigRaw(config)); len(errs) > 0 {
				t.Fatal(errs)
			}
			if _, err := r.Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
				t.Fatal(err)
			}
			d := schema.TestResourceDataRaw(t, r.Schema, config)
			cfg := ucloud.NewConfig()
			client := sdkuk8s.NewClient(&cfg, &auth.Credential{})
			req, err := buildUK8SClusterCreateRequest(client, d)
			if err != nil {
				t.Fatal(err)
			}
			form, err := request.EncodeForm(req)
			if err != nil {
				t.Fatal(err)
			}
			for key, want := range test.want {
				if form[key] != want {
					t.Errorf("%s = %q, want %q", key, form[key], want)
				}
			}
			for _, key := range test.absent {
				if _, ok := form[key]; ok {
					t.Errorf("unexpected field %s", key)
				}
			}
			if len(test.workers) == 0 {
				for key := range form {
					if strings.HasPrefix(key, "Nodes.") {
						t.Errorf("unexpected worker field %s", key)
					}
				}
			}
		})
	}
}

func TestClusterWorkerPlan(t *testing.T) {
	for _, test := range []struct {
		name        string
		change      map[string]interface{}
		remove      bool
		wantReplace bool
		wantError   bool
	}{
		{name: "unchanged"},
		{name: "count", change: map[string]interface{}{"count": 2}, wantReplace: true},
		{name: "CPU", change: map[string]interface{}{"cpu": 4}, wantReplace: true},
		{name: "image", change: map[string]interface{}{"image_id": "uimage-newimage"}, wantReplace: true},
		{name: "remove", remove: true, wantReplace: true},
		{name: "zero count", change: map[string]interface{}{"count": 0}, wantError: true},
		{name: "too many nodes", change: map[string]interface{}{"count": 11}, wantError: true},
		{name: "memory units", change: map[string]interface{}{"memory": 4}, wantError: true},
		{name: "data disk size", change: map[string]interface{}{"data_disk_size": 10}, wantError: true},
		{name: "family", change: map[string]interface{}{"uhost_family": "o2i"}, wantError: true},
		{name: "outstanding disks", change: map[string]interface{}{"machine_type": "O"}, wantError: true},
		{name: "network", change: map[string]interface{}{"net_capability": "invalid"}, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := resourceUCloudUK8SCluster()
			worker := map[string]interface{}{"availability_zone": "sg-02", "machine_type": "N", "cpu": 2, "memory": 4096}
			d := schema.TestResourceDataRaw(t, r.Schema, workerTestConfig(worker))
			d.SetId("uk8s-test")
			state := d.State()
			for key, value := range test.change {
				worker[key] = value
			}
			config := workerTestConfig(worker)
			if test.remove {
				config = workerTestConfig()
			}
			raw := terraform.NewResourceConfigRaw(config)
			_, errs := r.Validate(raw)
			if len(errs) > 0 {
				if !test.wantError {
					t.Fatal(errs)
				}
				return
			}
			diff, err := r.Diff(state, raw, nil)
			if test.wantError {
				if err == nil {
					t.Fatal("invalid configuration accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := diff != nil && diff.RequiresNew(); got != test.wantReplace {
				t.Fatalf("replacement = %v, want %v", got, test.wantReplace)
			}
			if !test.wantReplace && diff != nil && !diff.Empty() {
				t.Fatalf("unchanged configuration produced diff: %#v", diff)
			}
		})
	}
}

func TestClusterWorkerUnknownPlatform(t *testing.T) {
	r := resourceUCloudUK8SCluster()
	// Terraform SDK v1's representation of a value resolved only at apply.
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"
	config := workerTestConfig(map[string]interface{}{
		"availability_zone": "sg-02", "machine_type": "O", "cpu": 2, "memory": 4096,
		"uhost_family": "o2i", "boot_disk_type": "cloud_rssd", "min_cpu_platform": unknown,
	})
	if _, err := r.Diff(nil, terraform.NewResourceConfigRaw(config), nil); err != nil {
		t.Fatalf("unknown Worker platform must be deferred until apply: %v", err)
	}
}
