package ucloud

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceUCloudInstanceTypes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUCloudInstanceTypesRead,
		Schema: map[string]*schema.Schema{
			"cpu": &schema.Schema{
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateIntegerInRange(1, 32),
			},

			"memory": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateIntegerInRange(1, 128),
			},

			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"instance_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUCloudInstanceTypesRead(d *schema.ResourceData, meta interface{}) error {
	cpu := d.Get("cpu").(int)
	memory := d.Get("memory").(int)

	var instanceTypeIds []string
	var hostScaleType string
	if memory/cpu == 1 {
		hostScaleType = "highcpu"
	}

	if memory/cpu == 2 {
		hostScaleType = "basic"
	}

	if memory/cpu == 4 {
		hostScaleType = "standard"
	}

	if memory/cpu == 8 {
		hostScaleType = "highmem"
	}

	if hostScaleType != "" {
		normal := strings.Join([]string{"n", hostScaleType, strconv.Itoa(cpu)}, "-")
		instanceTypeIds = append(instanceTypeIds, normal)
	}

	customize := strings.Join([]string{"n", "customize", strconv.Itoa(cpu), strconv.Itoa(memory)}, "-")
	instanceTypeIds = append(instanceTypeIds, customize)

	err := dataSourceUCloudInstanceTypesSave(d, instanceTypeIds)
	if err != nil {
		return fmt.Errorf("error in read instance types, %s", err)
	}

	return nil
}

func dataSourceUCloudInstanceTypesSave(d *schema.ResourceData, instanceTypeIds []string) error {
	ids := []string{}
	data := []map[string]interface{}{}

	for _, item := range instanceTypeIds {
		ids = append(ids, item)
		data = append(data, map[string]interface{}{
			"id": item,
		})
	}

	d.SetId(hashStringArray(ids))
	if err := d.Set("instance_types", data); err != nil {
		return err
	}

	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
		writeToFile(outputFile.(string), data)
	}

	return nil
}
