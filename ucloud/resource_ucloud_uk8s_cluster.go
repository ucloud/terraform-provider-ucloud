package ucloud

import (
	"encoding/base64"
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/services/uk8s"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

func resourceUCloudUK8sCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceUCloudUK8sClusterCreate,
		Read:   resourceUCloudUK8sClusterRead,
		Update: resourceUCloudUK8sClusterUpdate,
		Delete: resourceUCloudUK8sClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

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

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateName,
			},

			//todo:review optional
			//"image_id": {
			//	Type:     schema.TypeString,
			//	Required: true,
			//	ForceNew: true,
			//},

			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ForceNew:     true,
				ValidateFunc: validateInstancePassword,
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

			//todo: to bool
			"enable_external_api_server": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"kube_proxy_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"delete_disks_with_instance": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"master": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 3,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"master_instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateInstanceType,
			},

			"master_boot_disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"local_normal",
					"local_ssd",
					"cloud_ssd",
					"cloud_rssd",
				}, false),
			},

			"master_data_disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validateAll(
					validation.IntBetween(20, 1000),
					validateMod(10),
				),
			},

			"master_data_disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"cloud_normal",
					"cloud_ssd",
					"cloud_rssd",
				}, false),
			},

			"master_min_cpu_platform": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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

			"nodes": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						//"count": {
						//	Type:     schema.TypeInt,
						//	Required: true,
						//	ForceNew: true,
						//	ValidateFunc: validation.IntBetween(1,10),
						//},
						//gpu

						//"node_id"

						"availability_zone": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"instance_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateInstanceType,
						},

						//todo
						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								if _, ok := v.(map[string]interface{}); ok {
									if len(v.(map[string]interface{})) > 5 {
										errors = append(errors, fmt.Errorf("only support five sets of labels for nodes"))
									}
								}
								return
							},
						},

						"max_pods": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},

						"isolation_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"boot_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"local_normal",
								"local_ssd",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"data_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: validateAll(
								validation.IntBetween(20, 1000),
								validateMod(10),
							),
						},

						"data_disk_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"cloud_normal",
								"cloud_ssd",
								"cloud_rssd",
							}, false),
						},

						"min_cpu_platform": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
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
		},
	}
}

func resourceUCloudUK8sClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn

	req := conn.NewCreateUK8SClusterV2Request()
	req.ServiceCIDR = ucloud.String(d.Get("service_cidr").(string))
	req.VPCId = ucloud.String(d.Get("vpc_id").(string))
	req.SubnetId = ucloud.String(d.Get("subnet_id").(string))

	if v, ok := d.GetOk("name"); ok {
		req.ClusterName = ucloud.String(v.(string))
	} else {
		req.ClusterName = ucloud.String(resource.PrefixedUniqueId("tf-uk8s-"))
	}
	//imageId := d.Get("image_id").(string)
	//req.ImageId = ucloud.String(imageId)
	req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))

	//imageResp, err := client.describeImageById(imageId)
	//if err != nil {
	//	return fmt.Errorf("error on reading image %q when creating uk8s cluster, %s", imageId, err)
	//}
	//todo
	//if v, ok := d.GetOk("user_data"); ok {
	//	if isStringIn("CloudInit", imageResp.Features) {
	//		req.UserData = ucloud.String(base64.StdEncoding.EncodeToString([]byte(v.(string))))
	//	} else {
	//		return fmt.Errorf("error on creating instance, the image %s must have %q feature, got %#v", imageId, "CloudInit", imageResp.Features)
	//	}
	//}
	if v, ok := d.GetOk("init_script"); ok {
		req.InitScript = ucloud.String(base64.StdEncoding.EncodeToString([]byte(v.(string))))
	}

	if v, ok := d.GetOk("charge_type"); ok {
		req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
	} else {
		req.ChargeType = ucloud.String("Month")
	}

	if v, ok := d.GetOkExists("duration"); ok {
		req.Quantity = ucloud.Int(v.(int))
	} else {
		req.Quantity = ucloud.Int(1)
	}

	if v, ok := d.GetOk("k8s_version"); ok {
		req.K8sVersion = ucloud.String(v.(string))
	}

	if v, ok := d.GetOk("enable_external_api_server"); ok {
		if v.(bool) {
			req.ExternalApiServer = ucloud.String("Yes")
		} else {
			req.ExternalApiServer = ucloud.String("No")
		}
	}

	if v, ok := d.GetOk("kube_proxy_mode"); ok {
		req.KubeProxy.Mode = ucloud.String(v.(string))
	}

	for _, item := range d.Get("master").([]interface{}) {
		master, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error on parse master when creating uk8s cluster")
		}
		masterNode := uk8s.CreateUK8SClusterV2ParamMaster{}
		masterNode.Zone = ucloud.String(master["availability_zone"].(string))
		req.Master = append(req.Master, masterNode)
	}

	// skip error because it has been validated by schema
	mt, _ := parseInstanceType(d.Get("master_instance_type").(string))
	req.MasterCPU = ucloud.Int(mt.CPU)
	req.MasterMem = ucloud.Int(mt.Memory)
	req.MasterMachineType = ucloud.String(strings.ToUpper(mt.HostType))

	if v, ok := d.GetOk("master_boot_disk_type"); ok {
		req.MasterBootDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
	} else {
		req.MasterBootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
	}

	if v, ok := d.GetOk("master_data_disk_size"); ok {
		if v, ok := d.GetOk("master_data_disk_type"); ok {
			req.MasterDataDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
		} else {
			req.MasterDataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
		}

		req.MasterDataDiskSize = ucloud.Int(v.(int))
	}

	if v, ok := d.GetOk("master_min_cpu_platform"); ok {
		req.MasterMinmalCpuPlatform = ucloud.String(v.(string))
	} else {
		req.MasterMinmalCpuPlatform = ucloud.String("Intel/Auto")
	}

	for _, item := range d.Get("nodes").([]interface{}) {
		node, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error on parse nodes when creating uk8s cluster")
		}
		nodeSet := uk8s.CreateUK8SClusterV2ParamNodes{}
		nodeSet.Zone = ucloud.String(node["availability_zone"].(string))
		nodeSet.Count = ucloud.Int(1)

		if v, ok := node["labels"]; ok {
			labels := make([]string, 0)
			labelMap := v.(map[string]interface{})
			for key, value := range labelMap {
				if value != nil {
					if vStr, ok := value.(string); ok {
						labels = append(labels, fmt.Sprintf("%s=%s", key, vStr))
					}
				}
			}

			nodeSet.Labels = ucloud.String(strings.Join(labels, ","))
		}

		if v, ok := node["max_pods"]; ok {
			nodeSet.MaxPods = ucloud.Int(v.(int))
		} else {
			nodeSet.MaxPods = ucloud.Int(110)
		}

		// skip error because it has been validated by schema
		nt, _ := parseInstanceType(node["instance_type"].(string))
		nodeSet.CPU = ucloud.Int(nt.CPU)
		nodeSet.Mem = ucloud.Int(nt.Memory)
		nodeSet.MachineType = ucloud.String(strings.ToUpper(nt.HostType))

		if v, ok := node["isolation_group"]; ok {
			nodeSet.IsolationGroup = ucloud.String(v.(string))
		}

		if v, ok := node["boot_disk_type"]; ok {
			nodeSet.BootDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
		} else {
			nodeSet.BootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
		}

		if v, ok := node["data_disk_size"]; ok {
			if v, ok := node["data_disk_type"]; ok {
				nodeSet.DataDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
			} else {
				nodeSet.DataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
			}

			nodeSet.DataDiskSize = ucloud.Int(v.(int))
		}

		if v, ok := node["min_cpu_platform"]; ok {
			nodeSet.MinmalCpuPlatform = ucloud.String(v.(string))
		}
		req.Nodes = append(req.Nodes, nodeSet)
	}

	resp, err := conn.CreateUK8SClusterV2(req)
	if err != nil {
		return fmt.Errorf("error on creating uk8s cluster, %s", err)
	}

	d.SetId(resp.ClusterId)

	// after create instance, we need to wait it initialized
	stateConf := &resource.StateChangeConf{
		Pending:    []string{statusPending},
		Target:     []string{statusRUNNING},
		Refresh:    uk8sClusterStateRefreshFunc(client, d.Id(), statusRUNNING),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      100 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error on waiting for uk8s cluster %q complete creating, %s", d.Id(), err)
	}

	return resourceUCloudUK8sClusterRead(d, meta)
}

func resourceUCloudUK8sClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn
	d.Partial(true)

	if d.HasChange("name") && !d.IsNewResource() {
		req := conn.NewGenericRequest()
		err := req.SetPayload(map[string]interface{}{
			"Action":      "ModifyUK8SClusterName",
			"ClusterId":   d.Id(),
			"ClusterName": d.Get("name").(string),
		})
		if err != nil {
			return fmt.Errorf("error on setting %s request to k8s cluster %q, %s", "ModifyUK8SClusterName", d.Id(), err)
		}
		_, err = conn.GenericInvoke(req)
		if err != nil {
			return fmt.Errorf("error on %s to k8s cluster %q, %s", "ModifyUK8SClusterName", d.Id(), err)
		}

		d.SetPartial("name")
	}

	if d.HasChange("nodes") && !d.IsNewResource() {
		o, n := d.GetChange("nodes")
		oldList := o.([]interface{})
		newList := n.([]interface{})
		notNeedUpdate := make([]interface{}, 0)
		needAdd := make([]interface{}, 0)
		needDelete := make([]interface{}, 0)
		for _, nv := range newList {
			for _, ov := range oldList {
				if reflect.DeepEqual(nv, ov) {
					notNeedUpdate = append(notNeedUpdate, nv)
					continue
				}
			}
			needAdd = append(needAdd, nv)
		}
		for _, ov := range oldList {
			for _, comm := range notNeedUpdate {
				if reflect.DeepEqual(ov, comm) {
					continue
				}
			}
			needDelete = append(needDelete, ov)
		}

		if len(needAdd) > 0 {
			for _, item := range needAdd {
				req := conn.NewAddUK8SUHostNodeRequest()
				node, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("error on parse node when %s to k8s cluster %q", "AddUK8SUHostNode", d.Id())
				}
				req.ClusterId = ucloud.String(d.Id())
				req.Zone = ucloud.String(node["availability_zone"].(string))
				req.Count = ucloud.Int(1)
				req.Password = ucloud.String(base64.StdEncoding.EncodeToString([]byte(d.Get("password").(string))))

				if v, ok := d.GetOk("charge_type"); ok {
					req.ChargeType = ucloud.String(upperCamelCvt.unconvert(v.(string)))
				} else {
					req.ChargeType = ucloud.String("Month")
				}

				if v, ok := d.GetOkExists("duration"); ok {
					req.Quantity = ucloud.Int(v.(int))
				} else {
					req.Quantity = ucloud.Int(1)
				}

				if v, ok := node["labels"]; ok {
					labels := make([]string, 0)
					labelMap := v.(map[string]interface{})
					for key, value := range labelMap {
						if value != nil {
							if vStr, ok := value.(string); ok {
								labels = append(labels, fmt.Sprintf("%s=%s", key, vStr))
							}
						}
					}

					req.Labels = ucloud.String(strings.Join(labels, ","))
				}
				// skip error because it has been validated by schema
				nt, _ := parseInstanceType(node["instance_type"].(string))
				req.CPU = ucloud.Int(nt.CPU)
				req.Mem = ucloud.Int(nt.Memory)
				req.MachineType = ucloud.String(strings.ToUpper(nt.HostType))

				if v, ok := node["isolation_group"]; ok {
					req.IsolationGroup = ucloud.String(v.(string))
				}

				if v, ok := node["boot_disk_type"]; ok {
					req.BootDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
				} else {
					req.BootDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
				}

				if v, ok := node["data_disk_size"]; ok {
					if v, ok := node["data_disk_type"]; ok {
						req.DataDiskType = ucloud.String(upperCvt.unconvert(v.(string)))
					} else {
						req.DataDiskType = ucloud.String(upperCvt.unconvert("cloud_ssd"))
					}

					req.DataDiskSize = ucloud.String(strconv.Itoa(v.(int)))
				}

				if v, ok := node["min_cpu_platform"]; ok {
					req.MinmalCpuPlatform = ucloud.String(v.(string))
				}
				if _, err := conn.AddUK8SUHostNode(req); err != nil {
					return fmt.Errorf("error on %s to uk8s cluster %q, %s", "AddUK8SUHostNode", d.Id(), err)
				}
			}
		}

		if len(needDelete) > 0 {
			for _, item := range needDelete {
				req := conn.NewDelUK8SClusterNodeV2Request()
				node, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("error on parse node when %s to k8s cluster %q", "DelUK8SClusterNodeV2", d.Id())
				}
				req.Zone = ucloud.String(node["availability_zone"].(string))
				req.ClusterId = ucloud.String(d.Id())
				//todo get nodeID
				//req.NodeId = ucloud.String(node["id"].(string))
				if v, ok := d.GetOkExists("delete_disks_with_instance"); ok {
					req.ReleaseDataUDisk = ucloud.Bool(v.(bool))
				}
				if _, err := conn.DelUK8SClusterNodeV2(req); err != nil {
					return fmt.Errorf("error on %s to uk8s cluster %q, %s", "DelUK8SClusterNodeV2", d.Id(), err)
				}
			}
		}

		d.SetPartial("nodes")
	}

	d.Partial(false)

	return resourceUCloudUK8sClusterRead(d, meta)
}

func resourceUCloudUK8sClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)

	instance, err := client.describeUK8sClusterById(d.Id())

	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading k8s cluster %q, %s", d.Id(), err)
	}

	//d.Set("password", d.Get("password"))
	d.Set("service_cidr", instance.ServiceCIDR)
	d.Set("name", instance.ClusterName)
	d.Set("vpc_id", instance.VPCId)
	d.Set("subnet_id", instance.SubnetId)
	d.Set("status", instance.Status)
	d.Set("create_time", timestampToString(instance.CreateTime))
	d.Set("api_server", instance.ApiServer)
	d.Set("external_api_server", instance.ExternalApiServer)

	d.Set("k8s_version", instance.K8sVersion)
	d.Set("pod_cidr", instance.PodCIDR)
	d.Set("external_api_server", instance.ExternalApiServer)

	nodeList, err := client.describeUK8sClusterNodeById(d.Id())
	if err != nil {
		if isNotFoundError(err) {
			//d.SetId("")
			return nil
		}
		return fmt.Errorf("error on reading uk8s cluster node list %q, %s", d.Id(), err)
	}
	nodes := []map[string]interface{}{}
	for _, node := range nodeList {
		if node.NodeRole == "node" {
			nodes = append(nodes, map[string]interface{}{
				"availability_zone": node.Zone,
				"instance_type":     instanceTypeSetFunc(upperCvt.convert(node.MachineType), node.CPU, node.Memory/1024),
				"status":            node.NodeStatus,
				"id":                node.NodeId,
			})
		}

	}

	return nil
}

func resourceUCloudUK8sClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*UCloudClient)
	conn := client.uk8sconn

	deleReq := conn.NewDelUK8SClusterRequest()
	deleReq.ClusterId = ucloud.String(d.Id())
	if v, ok := d.GetOkExists("delete_disks_with_instance"); ok {
		deleReq.ReleaseUDisk = ucloud.Bool(v.(bool))
	}

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		//instance, err := client.describeUK8sClusterById(d.Id())
		//if err != nil {
		//	if isNotFoundError(err) {
		//		return nil
		//	}
		//	return resource.NonRetryableError(fmt.Errorf("error on reading k8s cluster before deleting %q, %s", d.Id(), err))
		//}

		if _, err := conn.DelUK8SCluster(deleReq); err != nil {
			return resource.RetryableError(fmt.Errorf("error on deleting k8s cluster %q, %s", d.Id(), err))
		}

		// after create instance, we need to wait it initialized
		stateConf := &resource.StateChangeConf{
			Pending:    []string{statusPending},
			Target:     []string{statusDELETED},
			Refresh:    uk8sClusterStateRefreshFuncForDelete(client, d.Id(), statusDELETED),
			Timeout:    5 * time.Minute,
			Delay:      30 * time.Second,
			MinTimeout: 2 * time.Second,
		}

		_, err := stateConf.WaitForState()
		if err != nil {
			return resource.RetryableError(fmt.Errorf("error on waiting for uk8s cluster %q complete deleting, %s", d.Id(), err))
		}

		if _, err := client.describeUK8sClusterById(d.Id()); err != nil {
			if isNotFoundError(err) {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error on reading k8s cluster when deleting %q, %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("the specified k8s cluster %q has not been deleted due to unknown error", d.Id()))
	})
}

func uk8sClusterStateRefreshFuncForDelete(client *UCloudClient, instanceId, target string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := client.describeUK8sClusterById(instanceId)
		if err != nil {
			if isNotFoundError(err) {
				return statusDELETED, statusDELETED, nil
			}
			return nil, "", err
		}

		state := instance.Status
		if state != target {
			state = statusPending
		}

		return instance, state, nil
	}
}

func uk8sClusterStateRefreshFunc(client *UCloudClient, instanceId, target string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := client.describeUK8sClusterById(instanceId)
		if err != nil {
			if isNotFoundError(err) {
				return nil, statusPending, nil
			}
			return nil, "", err
		}

		state := instance.Status
		if state != target {
			state = statusPending
		}

		return instance, state, nil
	}
}
