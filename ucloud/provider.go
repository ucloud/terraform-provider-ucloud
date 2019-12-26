package ucloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
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
		},

		DataSourcesMap: map[string]*schema.Resource{
			"ucloud_projects":              dataSourceUCloudProjects(),
			"ucloud_images":                dataSourceUCloudImages(),
			"ucloud_zones":                 dataSourceUCloudZones(),
			"ucloud_eips":                  dataSourceUCloudEips(),
			"ucloud_instances":             dataSourceUCloudInstances(),
			"ucloud_lbs":                   dataSourceUCloudLBs(),
			"ucloud_lb_listeners":          dataSourceUCloudLBListeners(),
			"ucloud_lb_rules":              dataSourceUCloudLBRules(),
			"ucloud_lb_attachments":        dataSourceUCloudLBAttachments(),
			"ucloud_disks":                 dataSourceUCloudDisks(),
			"ucloud_db_instances":          dataSourceUCloudDBInstances(),
			"ucloud_security_groups":       dataSourceUCloudSecurityGroups(),
			"ucloud_subnets":               dataSourceUCloudSubnets(),
			"ucloud_lb_ssls":               dataSourceUCloudLBSSLs(),
			"ucloud_vpcs":                  dataSourceUCloudVPCs(),
			"ucloud_nat_gateways":          dateSourceUCloudNatGateways(),
			"ucloud_vpn_gateways":          dateSourceUCloudVPNGateways(),
			"ucloud_vpn_customer_gateways": dateSourceUCloudVPNCustomerGateways(),
			"ucloud_vpn_connections":       dateSourceUCloudVPNConnections(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"ucloud_instance":               resourceUCloudInstance(),
			"ucloud_eip":                    resourceUCloudEIP(),
			"ucloud_eip_association":        resourceUCloudEIPAssociation(),
			"ucloud_vpc":                    resourceUCloudVPC(),
			"ucloud_subnet":                 resourceUCloudSubnet(),
			"ucloud_vpc_peering_connection": resourceUCloudVPCPeeringConnection(),
			"ucloud_udpn_connection":        resourceUCloudUDPNConnection(),
			"ucloud_lb":                     resourceUCloudLB(),
			"ucloud_lb_listener":            resourceUCloudLBListener(),
			"ucloud_lb_attachment":          resourceUCloudLBAttachment(),
			"ucloud_lb_rule":                resourceUCloudLBRule(),
			"ucloud_disk":                   resourceUCloudDisk(),
			"ucloud_disk_attachment":        resourceUCloudDiskAttachment(),
			"ucloud_security_group":         resourceUCloudSecurityGroup(),
			"ucloud_lb_ssl":                 resourceUCloudLBSSL(),
			"ucloud_lb_ssl_attachment":      resourceUCloudLBSSLAttachment(),
			"ucloud_db_instance":            resourceUCloudDBInstance(),
			"ucloud_redis_instance":         resourceUCloudRedisInstance(),
			"ucloud_memcache_instance":      resourceUCloudMemcacheInstance(),
			"ucloud_isolation_group":        resourceUCloudIsolationGroup(),
			"ucloud_vip":                    resourceUCloudVIP(),
			"ucloud_nat_gateway":            resourceUCloudNatGateway(),
			"ucloud_nat_gateway_rule":       resourceUCloudNatGatewayRule(),
			"ucloud_vpn_gateway":            resourceUCloudVPNGateway(),
			"ucloud_vpn_customer_gateway":   resourceUCloudVPNCustomerGateway(),
			"ucloud_vpn_connection":         resourceUCloudVPNConnection(),
		},
		ConfigureFunc: providerConfigure,
	}
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

	client, err := config.Client()
	return client, err
}

var ucloudMutexKV = mutexkv.NewMutexKV()

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
