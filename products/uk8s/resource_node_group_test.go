package uk8s

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	sdkuk8s "github.com/ucloud/ucloud-sdk-go/services/uk8s"
)

func TestUK8SNodeGroupSchedulingRoundTrip(t *testing.T) {
	resource := resourceUCloudUK8SNodeGroup()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"labels": map[string]interface{}{
			"role":        "worker1",
			"environment": "test",
		},
		"taint": []interface{}{
			map[string]interface{}{
				"key": "dedicated", "value": "test", "effect": "NoSchedule",
			},
		},
	})

	labels, taints, err := expandUK8SNodeGroupScheduling(data)
	if err != nil {
		t.Fatalf("expand scheduling: %v", err)
	}
	if labels != "environment=test,role=worker1" {
		t.Fatalf("labels = %q", labels)
	}
	if taints != "dedicated=test:NoSchedule" {
		t.Fatalf("taints = %q", taints)
	}
	if got := flattenUK8SNodeGroupLabels(labels); !reflect.DeepEqual(got, map[string]interface{}{
		"environment": "test", "role": "worker1",
	}) {
		t.Fatalf("flatten labels = %#v", got)
	}
	wantTaints := []map[string]interface{}{{
		"key": "dedicated", "value": "test", "effect": "NoSchedule",
	}}
	if got := flattenUK8SNodeGroupTaints(taints); !reflect.DeepEqual(got, wantTaints) {
		t.Fatalf("flatten taints = %#v", got)
	}
}

func TestUK8SNodeGroupSchedulingRejectsDuplicateTaintKeys(t *testing.T) {
	resource := resourceUCloudUK8SNodeGroup()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"taint": []interface{}{
			map[string]interface{}{"key": "dedicated", "value": "a", "effect": "NoSchedule"},
			map[string]interface{}{"key": "dedicated", "value": "b", "effect": "NoExecute"},
		},
	})
	if _, _, err := expandUK8SNodeGroupScheduling(data); err == nil {
		t.Fatal("duplicate taint keys were accepted")
	}
}

func TestSetUK8SNodeGroupStateUsesMachineSpec(t *testing.T) {
	resource := resourceUCloudUK8SNodeGroup()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"cluster_id": "uk8s-test", "name": "workers", "availability_zone": "sg-02",
		"subnet_id": "subnet-test", "machine_type": "N", "cpu": 2, "memory": 4096,
	})
	setUK8SNodeGroupState(data, sdkuk8s.NodeGroupSet{
		MachineType: "O", CPU: 4, Mem: 8192,
	})
	if data.Get("machine_type") != "O" || data.Get("cpu") != 4 || data.Get("memory") != 8192 {
		t.Fatalf("node group machine spec = %q/%d/%d, want O/4/8192",
			data.Get("machine_type"), data.Get("cpu"), data.Get("memory"))
	}
}
