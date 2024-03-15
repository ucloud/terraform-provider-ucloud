package ucloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/mutexkv"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

			"assume_role": assumeRoleSchema(),
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
			"ucloud_db_parameter_groups":   dataSourceUCloudDBParameterGroups(),
			"ucloud_ufs_volumes":           dataSourceUCloudUFSVolumes(),
			"ucloud_us3_buckets":           dataSourceUCloudUS3Buckets(),
			"ucloud_db_backups":            dataSourceUCloudDBBackups(),
			"ucloud_anti_ddos_instances":   dataSourceUCloudAntiDDoSInstances(),
			"ucloud_anti_ddos_ips":         dataSourceUCloudAntiDDoSIPs(),
			"ucloud_iam_users":             dataSourceUCloudIAMUsers(),
			"ucloud_iam_groups":            dataSourceUCloudIAMGroups(),
			"ucloud_iam_projects":          dataSourceUCloudIAMProjects(),
			"ucloud_iam_policy":            dataSourceUCloudIAMPolicy(),
			"ucloud_iam_policy_document":   dataSourceUCloudIAMPolicyDocument(),
			"ucloud_baremetal_images":      dataSourceUCloudBareMetalImages(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"ucloud_instance":                    resourceUCloudInstance(),
			"ucloud_eip":                         resourceUCloudEIP(),
			"ucloud_eip_association":             resourceUCloudEIPAssociation(),
			"ucloud_vpc":                         resourceUCloudVPC(),
			"ucloud_subnet":                      resourceUCloudSubnet(),
			"ucloud_vpc_peering_connection":      resourceUCloudVPCPeeringConnection(),
			"ucloud_udpn_connection":             resourceUCloudUDPNConnection(),
			"ucloud_lb":                          resourceUCloudLB(),
			"ucloud_lb_listener":                 resourceUCloudLBListener(),
			"ucloud_lb_attachment":               resourceUCloudLBAttachment(),
			"ucloud_lb_rule":                     resourceUCloudLBRule(),
			"ucloud_disk":                        resourceUCloudDisk(),
			"ucloud_disk_attachment":             resourceUCloudDiskAttachment(),
			"ucloud_security_group":              resourceUCloudSecurityGroup(),
			"ucloud_lb_ssl":                      resourceUCloudLBSSL(),
			"ucloud_lb_ssl_attachment":           resourceUCloudLBSSLAttachment(),
			"ucloud_db_instance":                 resourceUCloudDBInstance(),
			"ucloud_redis_instance":              resourceUCloudRedisInstance(),
			"ucloud_memcache_instance":           resourceUCloudMemcacheInstance(),
			"ucloud_isolation_group":             resourceUCloudIsolationGroup(),
			"ucloud_vip":                         resourceUCloudVIP(),
			"ucloud_nat_gateway":                 resourceUCloudNatGateway(),
			"ucloud_nat_gateway_rule":            resourceUCloudNatGatewayRule(),
			"ucloud_vpn_gateway":                 resourceUCloudVPNGateway(),
			"ucloud_vpn_customer_gateway":        resourceUCloudVPNCustomerGateway(),
			"ucloud_vpn_connection":              resourceUCloudVPNConnection(),
			"ucloud_ufs_volume":                  resourceUCloudUFSVolume(),
			"ucloud_ufs_volume_mount_point":      resourceUCloudUFSVolumeMountPoint(),
			"ucloud_us3_bucket":                  resourceUCloudUS3Bucket(),
			"ucloud_uk8s_cluster":                resourceUCloudUK8SCluster(),
			"ucloud_uk8s_node":                   resourceUCloudUK8SNode(),
			"ucloud_anti_ddos_instance":          resourceUCloudAntiDDoSInstance(),
			"ucloud_anti_ddos_allowed_domain":    resourceUCloudAntiDDoSAllowedDomain(),
			"ucloud_anti_ddos_ip":                resourceUCloudAntiDDoSIP(),
			"ucloud_anti_ddos_rule":              resourceUCloudAntiDDoSRule(),
			"ucloud_iam_access_key":              resourceUCloudIAMAccessKey(),
			"ucloud_iam_user":                    resourceUCloudIAMUser(),
			"ucloud_iam_group":                   resourceUCloudIAMGroup(),
			"ucloud_iam_group_membership":        resourceUCloudIAMGroupMembership(),
			"ucloud_iam_project":                 resourceUCloudIAMProject(),
			"ucloud_iam_policy":                  resourceUCloudIAMPolicy(),
			"ucloud_iam_user_policy_attachment":  resourceUCloudIAMUserPolicyAttachment(),
			"ucloud_iam_group_policy_attachment": resourceUCloudIAMGroupPolicyAttachment(),
			"ucloud_instance_state":              resourceUCloudInstanceState(),
			"ucloud_baremetal_instance":          resourceUCloudBareMetalInstance(),
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
