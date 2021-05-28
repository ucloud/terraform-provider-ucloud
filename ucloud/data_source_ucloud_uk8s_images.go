package ucloud

//func dataSourceUCloudUK8sImages() *schema.Resource {
//	return &schema.Resource{
//		Read: dataSourceUCloudUK8sImagesRead,
//		Schema: map[string]*schema.Schema{
//			"availability_zone": {
//				Type:     schema.TypeString,
//				Optional: true,
//			},
//
//			"name_regex": {
//				Type:         schema.TypeString,
//				Optional:     true,
//				ValidateFunc: validation.ValidateRegexp,
//			},
//
//			"ids": {
//				Type:     schema.TypeSet,
//				Optional: true,
//				Elem: &schema.Schema{
//					Type: schema.TypeString,
//				},
//				Set:           schema.HashString,
//				Computed:      true,
//			},
//
//			"output_file": {
//				Type:     schema.TypeString,
//				Optional: true,
//			},
//
//			"total_count": {
//				Type:     schema.TypeInt,
//				Computed: true,
//			},
//
//			"uk8s_images": {
//				Type:     schema.TypeList,
//				Computed: true,
//				Elem: &schema.Resource{
//					Schema: map[string]*schema.Schema{
//						"id": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"name": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"type": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"size": {
//							Type:     schema.TypeInt,
//							Computed: true,
//						},
//
//						"availability_zone": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"os_type": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"os_name": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"features": {
//							Type: schema.TypeList,
//							Elem: &schema.Schema{
//								Type: schema.TypeString,
//							},
//							Computed: true,
//						},
//
//						"create_time": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"description": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//
//						"status": {
//							Type:     schema.TypeString,
//							Computed: true,
//						},
//					},
//				},
//			},
//		},
//	}
//}
//
//func dataSourceUCloudUK8sImagesRead(d *schema.ResourceData, meta interface{}) error {
//	conn := meta.(*UCloudClient).uk8sconn
//
//	req := conn.NewDescribeUK8SImageRequest()
//
//	if v, ok := d.GetOk("availability_zone"); ok {
//		req.Zone = ucloud.String(v.(string))
//	}
//	var allImages []uhost.UHostImageSet
//	var offset int
//	limit := 100
//	for {
//		req.Limit = ucloud.Int(limit)
//		req.Offset = ucloud.Int(offset)
//		resp, err := conn.DescribeUK8SImage(req)
//		if err != nil {
//			return fmt.Errorf("error on reading image list, %s", err)
//		}
//
//		if resp == nil || len(resp.ImageSet) < 1 {
//			break
//		}
//
//		for _, v := range resp.ImageSet {
//			if v.State == "Available" {
//				allImages = append(allImages, v)
//			}
//		}
//
//		if len(resp.ImageSet) < limit {
//			break
//		}
//
//		offset = offset + limit
//	}
//	ids, idsOk := d.GetOk("ids")
//	nameRegex, nameRegexOk := d.GetOk("name_regex")
//
//	var filteredImages []uhost.UHostImageSet
//
//	if idsOk || nameRegexOk {
//		var r *regexp.Regexp
//		if nameRegex != "" {
//			r = regexp.MustCompile(nameRegex.(string))
//		}
//		for _, image := range allImages {
//			if r != nil && !r.MatchString(image.ImageName) {
//				continue
//			}
//
//			if idsOk && !isStringIn(image.ImageId, schemaSetToStringSlice(ids)) {
//				continue
//			}
//			filteredImages = append(filteredImages, image)
//		}
//	} else {
//		filteredImages = allImages
//	}
//
//	var finalImages []uhost.UHostImageSet
//	if len(filteredImages) > 1 && d.Get("most_recent").(bool) {
//		sort.Slice(filteredImages, func(i, j int) bool {
//			return int64(filteredImages[i].CreateTime) > int64(filteredImages[j].CreateTime)
//		})
//
//		finalImages = []uhost.UHostImageSet{filteredImages[0]}
//	} else {
//		finalImages = filteredImages
//	}
//
//	err := dataSourceUCloudImagesSave(d, finalImages)
//	if err != nil {
//		return fmt.Errorf("error on reading image list, %s", err)
//	}
//
//	return nil
//}
//
//func dataSourceUCloudImagesSave(d *schema.ResourceData, projects []uhost.UHostImageSet) error {
//	ids := []string{}
//	data := []map[string]interface{}{}
//
//	for _, item := range projects {
//		ids = append(ids, item.ImageId)
//		data = append(data, map[string]interface{}{
//			"id":                item.ImageId,
//			"name":              item.ImageName,
//			"availability_zone": item.Zone,
//			"type":              upperCamelCvt.convert(item.ImageType),
//			"os_type":           upperCamelCvt.convert(item.OsType),
//			"os_name":           item.OsName,
//			"features":          item.Features,
//			"create_time":       timestampToString(item.CreateTime),
//			"size":              item.ImageSize,
//			"description":       item.ImageDescription,
//			"status":            item.State,
//		})
//	}
//
//	d.SetId(hashStringArray(ids))
//	d.Set("total_count", len(data))
//	d.Set("ids", ids)
//	if err := d.Set("images", data); err != nil {
//		return err
//	}
//
//	if outputFile, ok := d.GetOk("output_file"); ok && outputFile.(string) != "" {
//		writeToFile(outputFile.(string), data)
//	}
//
//	return nil
//}
