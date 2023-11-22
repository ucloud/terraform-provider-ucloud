package ucloud

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceUCloudIAMPolicyDocument() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudIAMPolicyDocumentRead,
		Schema: map[string]*schema.Schema{
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "1",
				ValidateFunc: validation.IntInSlice([]int{1}),
			},
			"statement": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"effect": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Allow",
							ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
						},
						"action": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"resource": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func dataSourceUCloudIAMPolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	if v, ok := d.GetOk("statement"); ok {
		doc, err := assembleDataSourcePolicyJSON(v.([]interface{}), d.Get("version").(int))
		if err != nil {
			return err
		}

		if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
			writeToFile(output.(string), doc)
		}

		d.Set("json", doc)

		d.SetId(hashString(doc))
	}

	return nil
}

func assembleDataSourcePolicyJSON(statements []interface{}, version int) (string, error) {
	document := PolicyDocument{Version: version}
	for _, v := range statements {
		var statement PolicyStatement
		statementMap := v.(map[string]interface{})
		if action, ok := statementMap["action"]; ok {
			statement.Action = schemaListToStringSlice(action)
		}
		if resource, ok := statementMap["resource"]; ok {
			statement.Resource = schemaListToStringSlice(resource)
		}
		statement.Effect = statementMap["effect"].(string)
		document.Statement = append(document.Statement, statement)
	}
	jsonBytes, err := json.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("fail to marshal document: %v", err)
	}
	return string(jsonBytes), nil
}

type PolicyStatement struct {
	Action   []string `json:"Action"`
	Effect   string   `json:"Effect"`
	Resource []string `json:"Resource"`
}
type PolicyDocument struct {
	Version   int               `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
}
