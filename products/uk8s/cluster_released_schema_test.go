package uk8s

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"time"
)

// Frozen schema from 56da6bd15ecf08db7e59de0c4348d284250fe475.
// Keep independent of the current schema so upgrade tests detect regressions.
func testUK8SClusterReleasedSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUK8SClusterCreate,
		Read:   resourceUCloudUK8SClusterRead,
		Update: resourceUCloudUK8SClusterUpdate,
		Delete: resourceUCloudUK8SClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			diffValidateBootDiskTypeWithInstanceTypeOfUK8sCluster,
		),

		Schema: map[string]*schema.Schema{
			"service_cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validateInstancePassword,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			"user_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"init_script": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"charge_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"year",
					"month",
					"dynamic",
				}, false),
			},

			"duration": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateDuration,
			},

			"k8s_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"enable_external_api_server": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"delete_disks_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"kube_proxy": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"master": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							MinItems: 3,
							MaxItems: 3,
							Required: true,
							ForceNew: true,
						},

						"instance_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateInstanceType,
						},

						"boot_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local_normal",
								"local_ssd",
								"cloud_normal",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"data_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validateAll(
								validation.IntBetween(20, 1000),
								validateMod(10),
							),
						},

						"data_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local_normal",
								"local_ssd",
								"cloud_normal",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"min_cpu_platform": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "Intel/Auto",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return old == "" && new == "Intel/Auto"
							},
							ValidateFunc: validation.StringInSlice([]string{
								"Intel/Auto",
								"Intel/IvyBridge",
								"Intel/Haswell",
								"Intel/Broadwell",
								"Intel/Skylake",
								"Intel/Cascadelake",
								"Intel/CascadelakeR",
								"Amd/Auto",
								"Amd/Epyc2",
								"Ampere/Altra",
							}, false),
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"api_server": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"external_api_server": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"pod_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateUImageName,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					o, _ := d.GetChange("image_id")
					return o != ""
				},
			},
		},
	}
}
