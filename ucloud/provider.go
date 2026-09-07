package ucloud

import (
	"time"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/productcatalog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_PUBLIC_KEY", nil),
				Description: descriptions["public_key"],
			},

			"private_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_PRIVATE_KEY", nil),
				Description: descriptions["private_key"],
			},

			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_PROFILE", nil),
				Description: descriptions["profile"],
			},

			"shared_credentials_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_SHARED_CREDENTIAL_FILE", nil),
				Description: descriptions["shared_credentials_file"],
			},

			"region": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_REGION", nil),
				Description: descriptions["region"],
			},

			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_PROJECT_ID", nil),
				Description: descriptions["project_id"],
			},

			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     defaultMaxRetries,
				Description: descriptions["max_retries"],
			},

			"insecure": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   descriptions["insecure"],
				ConflictsWith: []string{"base_url"},
			},

			"base_url": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   descriptions["base_url"],
				ConflictsWith: []string{"insecure"},
				ValidateFunc:  validateBaseUrl,
			},

			"assume_role": assumeRoleSchema(),
		},

		DataSourcesMap: map[string]*schema.Resource{},
		ResourcesMap:   map[string]*schema.Resource{},
		ConfigureFunc:  providerConfigure,
	}
	product.MustRegister(provider, productcatalog.Bindings()...)
	return provider
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		PublicKey:             d.Get("public_key").(string),
		PrivateKey:            d.Get("private_key").(string),
		Region:                d.Get("region").(string),
		MaxRetries:            d.Get("max_retries").(int),
		Insecure:              d.Get("insecure").(bool),
		Profile:               d.Get("profile").(string),
		SharedCredentialsFile: d.Get("shared_credentials_file").(string),
	}

	if projectId, ok := d.GetOk("project_id"); ok && projectId.(string) != "" {
		config.ProjectId = projectId.(string)
	}

	// if no base url be set, get insecure http or secure https default url
	// if base url is set, use it
	if v, ok := d.GetOk("base_url"); ok && v.(string) != "" {
		config.BaseURL = v.(string)
	} else if config.Insecure {
		config.BaseURL = GetInsecureEndpointURL(config.Region)
	} else if !config.Insecure {
		config.BaseURL = GetEndpointURL(config.Region)
	}

	if v, ok := d.GetOk("assume_role"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.AssumeRole = expandAssumeRole(v.([]interface{})[0].(map[string]interface{}))
	}

	client, err := config.Client()
	return client, err
}

func assumeRoleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "The duration of the role session. Valid time units are ns, us (or µs), ms, s, h, or m.",
					ValidateFunc: validateAssumeRoleDuration,
					Default:      "900s",
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.ValidateJsonString,
				},
				"role_urn": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "UCloud Resource Name (URN) of an IAM Role to assume prior to making API calls.",
				},
				"session_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "An identifier for the assumed role session.",
				},
			},
		},
	}
}

func expandAssumeRole(tfMap map[string]interface{}) *AssumeRoleConfig {
	if tfMap == nil {
		return nil
	}

	assumeRole := AssumeRoleConfig{}

	if v, ok := tfMap["duration"].(string); ok && v != "" {
		duration, _ := time.ParseDuration(v)
		assumeRole.Duration = duration
	}

	if v, ok := tfMap["policy"].(string); ok && v != "" {
		assumeRole.Policy = v
	}

	if v, ok := tfMap["role_urn"].(string); ok && v != "" {
		assumeRole.RoleURN = v
	}

	if v, ok := tfMap["session_name"].(string); ok && v != "" {
		assumeRole.SessionName = v
	}

	return &assumeRole
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"public_key":              "...",
		"private_key":             "...",
		"region":                  "...",
		"project_id":              "...",
		"max_retries":             "...",
		"insecure":                "...",
		"base_url":                "...",
		"profile":                 "...",
		"shared_credentials_file": "...",
	}
}
