package uk8s

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func TestFormatUK8SInstanceType(t *testing.T) {
	for _, test := range []struct {
		machine string
		cpu     int
		memory  int
		want    string
	}{
		{machine: "O", cpu: 2, memory: 4096, want: "o-basic-2"},
		{machine: "N", cpu: 2, memory: 6144, want: "n-customized-2-6"},
	} {
		if got := formatUK8SInstanceType(test.machine, test.cpu, test.memory); got != test.want {
			t.Errorf("formatUK8SInstanceType(%q, %d, %d) = %q, want %q", test.machine, test.cpu, test.memory, got, test.want)
		}
	}
}
