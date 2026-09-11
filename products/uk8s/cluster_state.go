package uk8s

import "github.com/zclconf/go-cty/cty"

// uk8sClusterStateV0Type is frozen to the published version 0 schema.
// Do not derive historical state types from the current resource schema.
func uk8sClusterStateV0Type() cty.Type {
	return cty.Object(map[string]cty.Type{
		"api_server":                 cty.String,
		"charge_type":                cty.String,
		"create_time":                cty.String,
		"delete_disks_with_instance": cty.Bool,
		"duration":                   cty.Number,
		"enable_external_api_server": cty.Bool,
		"external_api_server":        cty.String,
		"image_id":                   cty.String,
		"init_script":                cty.String,
		"k8s_version":                cty.String,
		"kube_proxy": cty.List(cty.Object(map[string]cty.Type{
			"mode": cty.String,
		})),
		"master": cty.List(cty.Object(map[string]cty.Type{
			"availability_zones": cty.List(cty.String),
			"boot_disk_type":     cty.String,
			"data_disk_size":     cty.Number,
			"data_disk_type":     cty.String,
			"instance_type":      cty.String,
			"min_cpu_platform":   cty.String,
		})),
		"name":         cty.String,
		"password":     cty.String,
		"pod_cidr":     cty.String,
		"service_cidr": cty.String,
		"status":       cty.String,
		"subnet_id":    cty.String,
		"user_data":    cty.String,
		"vpc_id":       cty.String,
		"id":           cty.String,
		"timeouts":     cty.Object(map[string]cty.Type{"create": cty.String, "update": cty.String, "delete": cty.String}),
	})
}

// Version 1 makes instance_type optional and adds independent specifications.
// Preserve legacy values verbatim: no field is removed or reinterpreted, and
// existing state must not be subjected to new configuration validators.
func upgradeUK8SClusterStateV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	return rawState, nil
}
