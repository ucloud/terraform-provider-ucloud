package ucloud

import (
	"github.com/hashicorp/terraform/helper/mutexkv"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"public_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_PUBLIC_KEY", nil),
				Description: descriptions["public_key"],
			},

			"private_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("UCLOUD_PRIVATE_KEY", nil),
				Description: descriptions["private_key"],
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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     defaultInSecure,
				Description: descriptions["insecure"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"ucloud_projects": dataSourceUCloudProjects(),
			"ucloud_images":   dataSourceUCloudImages(),
			"ucloud_zones":    dataSourceUCloudZones(),
			"ucloud_eips":     dataSourceUCloudEips(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"ucloud_instance":               resourceUCloudInstance(),
			"ucloud_eip":                    resourceUCloudEIP(),
			"ucloud_eip_association":        resourceUCloudEIPAssociation(),
			"ucloud_vpc":                    resourceUCloudVPC(),
			"ucloud_subnet":                 resourceUCloudSubnet(),
			"ucloud_vpc_peering_connection": resourceUCloudVPCPeeringConnection(),
			"ucloud_lb":                     resourceUCloudLB(),
			"ucloud_lb_listener":            resourceUCloudLBListener(),
			"ucloud_lb_attachment":          resourceUCloudLBAttachment(),
			"ucloud_lb_rule":                resourceUCloudLBRule(),
			"ucloud_disk":                   resourceUCloudDisk(),
			"ucloud_disk_attachment":        resourceUCloudDiskAttachment(),
			"ucloud_security_group":         resourceUCloudSecurityGroup(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		PublicKey:  d.Get("public_key").(string),
		PrivateKey: d.Get("private_key").(string),
		Region:     d.Get("region").(string),
		MaxRetries: d.Get("max_retries").(int),
		Insecure:   d.Get("insecure").(bool),
	}

	if projectId, ok := d.GetOk("project_id"); ok && projectId.(string) != "" {
		config.ProjectId = projectId.(string)
	}

	client, err := config.Client()
	return client, err
}

var ucloudMutexKV = mutexkv.NewMutexKV()

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"public_key":  "...",
		"private_key": "...",
		"region":      "...",
		"project_id":  "...",
		"max_retries": "...",
		"insecure":    "...",
	}
}
