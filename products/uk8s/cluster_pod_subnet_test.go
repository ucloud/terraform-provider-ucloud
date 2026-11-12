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

func TestClusterPodSubnetCreateRequest(t *testing.T) {
	for _, test := range []struct {
		name   string
		config map[string]interface{}
		want   map[string]string
	}{
		{
			name: "omitted preserves API defaults",
		},
		{
			name:   "explicit false is sent",
			config: map[string]interface{}{"dedicated_pod_subnet": false},
			want:   map[string]string{"DedicatedPodSubnet": "false"},
		},
		{
			name: "dedicated subnet with firewall mode",
			config: map[string]interface{}{
				"dedicated_pod_subnet": true,
				"pod_subnet_id":        "subnet-pods",
			},
			want: map[string]string{"DedicatedPodSubnet": "true", "PodSubnetId": "subnet-pods"},
		},
		{
			name: "dedicated subnet with security group",
			config: map[string]interface{}{
				"dedicated_pod_subnet": true,
				"pod_subnet_id":        "subnet-pods",
				"pod_subnet_sec_group": "secgroup-pods",
			},
			want: map[string]string{
				"DedicatedPodSubnet": "true",
				"PodSubnetId":        "subnet-pods",
				"PodSubnetSecGroup":  "secgroup-pods",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := resourceUCloudUK8SCluster()
			config := masterTestConfig(map[string]interface{}{"instance_type": "n-basic-2"})
			for key, value := range test.config {
				config[key] = value
			}
			if _, errs := r.Validate(terraform.NewResourceConfigRaw(config)); len(errs) > 0 {
				t.Fatalf("validate: %v", errs)
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
			for _, key := range []string{"DedicatedPodSubnet", "PodSubnetId", "PodSubnetSecGroup"} {
				got, present := form[key]
				want, expected := test.want[key]
				if present != expected || got != want {
					t.Errorf("%s = %q (present %t), want %q (present %t)", key, got, present, want, expected)
				}
			}
		})
	}
}
