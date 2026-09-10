package uk8s

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type nodeGroupTestRuntime struct{ client *sdkuk8s.UK8SClient }

func (r nodeGroupTestRuntime) ProductClient(_ string, _ product.ClientConstructor) (interface{}, error) {
	return r.client, nil
}

type uk8sTestTransport func(*http.Request) (*http.Response, error)

func (f uk8sTestTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestNodeGroupImportMatchesCreate(t *testing.T) {
	for _, disk := range []string{"", "CLOUD_RSSD"} {
		t.Run("data-disk-"+disk, func(t *testing.T) {
			cfg := ucloud.NewConfig()
			cfg.Region, cfg.ProjectId = "sg", "org-test"
			client := sdkuk8s.NewClient(&cfg, &auth.Credential{PublicKey: "test", PrivateKey: "test"})
			client.SetTransport(uk8sTestTransport(func(req *http.Request) (*http.Response, error) {
				if err := req.ParseForm(); err != nil {
					return nil, err
				}
				action := req.Form.Get("Action")
				result := map[string]interface{}{"RetCode": 0, "Action": action + "Response"}
				switch action {
				case "AddUK8SNodeGroup":
					result["NodeGroupId"] = "ng-test"
				case "ListUK8SNodeGroup":
					if req.Form.Get("ClusterId") != "uk8s-test" {
						return nil, fmt.Errorf("wrong cluster in import read: %q", req.Form.Get("ClusterId"))
					}
					result["NodeGroupList"] = []map[string]interface{}{{
						"NodeGroupId": "ng-test", "NodeGroupName": "workers", "Zone": "sg-02",
						"SubnetId": "subnet-test", "MachineType": "N", "CPU": 2, "Mem": 4096,
						"ChargeType": "Dynamic", "BootDiskType": "CLOUD_SSD", "BootDiskSize": 40,
						"DataDiskType": disk, "DataDiskSize": 0, "MinimalCpuPlatform": "Intel/Auto",
						"MaxPods": 110, "Labels": "env=test", "NodeList": []string{},
					}}
				default:
					return nil, fmt.Errorf("unexpected API action %q", action)
				}
				body, err := json.Marshal(result)
				if err != nil {
					return nil, err
				}
				return &http.Response{
					StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}},
					Body: io.NopCloser(strings.NewReader(string(body))), Request: req,
				}, nil
			}))
			r := resourceUCloudUK8SNodeGroup()
			config := map[string]interface{}{
				"cluster_id": "uk8s-test", "name": "workers", "availability_zone": "sg-02",
				"subnet_id": "subnet-test", "instance_type": "n-basic-2", "charge_type": "dynamic",
				"labels": map[string]interface{}{"env": "test"},
			}
			if disk != "" {
				config["data_disk_type"] = "cloud_rssd"
			}
			created := schema.TestResourceDataRaw(t, r.Schema, config)
			runtime := nodeGroupTestRuntime{client: client}
			if err := r.Create(created, runtime); err != nil {
				t.Fatal(err)
			}
			imported := schema.TestResourceDataRaw(t, r.Schema, nil)
			imported.SetId("uk8s-test/ng-test")
			items, err := r.Importer.State(imported, runtime)
			if err != nil || len(items) != 1 {
				t.Fatalf("import returned %d items: %v", len(items), err)
			}
			if err := r.Read(imported, runtime); err != nil {
				t.Fatal(err)
			}
			if imported.Id() != created.Id() || !reflect.DeepEqual(imported.State().Attributes, created.State().Attributes) {
				t.Fatalf("import state differs from create state:\nimport: %#v\ncreate: %#v",
					imported.State().Attributes, created.State().Attributes)
			}
			plan, err := r.Diff(imported.State(), terraform.NewResourceConfigRaw(config), nil)
			if err != nil || (plan != nil && !plan.Empty()) {
				t.Fatalf("import produced an unexpected plan: %#v, %v", plan, err)
			}
		})
	}
}

func TestNodeGroupImportInvalidID(t *testing.T) {
	for _, id := range []string{"", "ng-test", "/ng-test", "uk8s-test/", "uk8s-test/ng-test/extra"} {
		t.Run(id, func(t *testing.T) {
			r := resourceUCloudUK8SNodeGroup()
			d := schema.TestResourceDataRaw(t, r.Schema, nil)
			d.SetId(id)
			if _, err := r.Importer.State(d, nil); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("%q", id)) {
				t.Fatalf("expected import error containing ID %q, got %v", id, err)
			}
			if d.Id() != id {
				t.Fatal("failed import changed the ID")
			}
		})
	}
}
