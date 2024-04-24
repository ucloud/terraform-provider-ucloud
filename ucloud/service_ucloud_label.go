package ucloud

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

const CustomLabelCategory = "custom"

func (client *UCloudClient) describeLabel(key, value string) (*label.ListLabelsLabel, error) {
	conn := client.labelconn

	limit := 100
	offset := 0
	for {
		req := conn.NewListLabelsRequest()
		req.Category = ucloud.String(CustomLabelCategory)
		resp, err := client.labelconn.ListLabels(req)
		if err != nil {
			return nil, err
		}
		if len(resp.Labels) < 1 {
			break
		}
		for _, label := range resp.Labels {
			if label.Key == key && label.Value == value {
				return &label, nil
			}
		}
		if len(resp.Labels) < limit {
			break
		}
		offset = offset + limit
	}
	return nil, newNotFoundError(getNotFoundMessage("label", buildUCloudLabelID(key, value)))

}

func (client *UCloudClient) describeLabelAttachment(key, value, resource string) (*label.ListResourcesByLabelsResource, error) {
	conn := client.labelconn
	limit := 100
	offset := 0
	listProjectsReq := conn.NewListProjectsByLabelsRequest()
	listProjectsReq.Labels = []label.ListProjectsByLabelsParamLabels{{Key: ucloud.String(key), Value: ucloud.String(value)}}
	listProjectsResp, err := conn.ListProjectsByLabels(listProjectsReq)
	if err != nil {
		return nil, fmt.Errorf("error on listing projects by labels, %s", err)
	}
	projectIds := make([]string, 0)
	resourceTypes := make([]string, 0)
	for _, project := range listProjectsResp.Projects {
		projectIds = append(projectIds, project.ProjectId)
		resourceTypes = append(resourceTypes, project.ResourceTypes...)
	}
	if len(projectIds) > 0 && len(resourceTypes) > 0 {
		for {
			req := conn.NewListResourcesByLabelsRequest()
			req.Limit = ucloud.Int(limit)
			req.Offset = ucloud.Int(offset)
			req.Labels = []label.ListResourcesByLabelsParamLabels{{Key: ucloud.String(key), Value: ucloud.String(value)}}

			req.ProjectIds = projectIds
			req.ResourceTypes = resourceTypes
			resp, err := conn.ListResourcesByLabels(req)
			if err != nil {
				return nil, fmt.Errorf("error on listing resources by labels, %s", err)
			}
			for _, resourceInfo := range resp.Resources {
				if resourceInfo.ResourceId == resource {
					return &resourceInfo, nil
				}
			}
			if len(resp.Resources) < limit {
				break
			}
			offset = offset + limit
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("label attachment", buildUCloudLabelAttachmentID(key, value, resource)))
}
