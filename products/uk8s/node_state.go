package uk8s

import "github.com/zclconf/go-cty/cty"

// uk8sNodeStateV0Type is frozen before independent node specifications were added.
// Do not derive historical state types from the current resource schema.
func uk8sNodeStateV0Type() cty.Type {
	return cty.Object(map[string]cty.Type{
		"availability_zone":          cty.String,
		"cluster_id":                 cty.String,
		"node_group_id":              cty.String,
		"image_id":                   cty.String,
		"password":                   cty.String,
		"instance_type":              cty.String,
		"uhost_family":               cty.String,
		"charge_type":                cty.String,
		"duration":                   cty.Number,
		"boot_disk_type":             cty.String,
		"data_disk_size":             cty.Number,
		"data_disk_type":             cty.String,
		"isolation_group":            cty.String,
		"subnet_id":                  cty.String,
		"user_data":                  cty.String,
		"init_script":                cty.String,
		"delete_disks_with_instance": cty.Bool,
		"disable_schedule_on_create": cty.Bool,
		"min_cpu_platform":           cty.String,
		"status":                     cty.String,
		"ip_set": cty.List(cty.Object(map[string]cty.Type{
			"ip":            cty.String,
			"internet_type": cty.String,
		})),
		"create_time": cty.String,
		"expire_time": cty.String,
		"id":          cty.String,
		"timeouts":    cty.Object(map[string]cty.Type{"create": cty.String, "update": cty.String, "delete": cty.String}),
	})
}

// Keep legacy values verbatim; do not fill new specification fields, which would
// conflict with instance_type and introduce a replacement on the next plan.
func upgradeUK8SNodeStateV0(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	return rawState, nil
}
