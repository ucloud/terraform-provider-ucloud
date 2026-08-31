package label

import (
	"fmt"

	labelapi "github.com/ucloud/ucloud-sdk-go/services/label"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

const CustomLabelCategory = "custom"

func describeLabel(client *labelapi.LabelClient, key, value string) (*labelapi.ListLabelsLabel, error) {
	limit := 100
	offset := 0
	for {
		req := client.NewListLabelsRequest()
		req.Category = ucloud.String(CustomLabelCategory)
		resp, err := client.ListLabels(req)
		if err != nil {
			return nil, err
		}
		if len(resp.Labels) < 1 {
			break
		}
		for _, item := range resp.Labels {
			if item.Key == key && item.Value == value {
				return &item, nil
			}
		}
		if len(resp.Labels) < limit {
			break
		}
		offset += limit
	}
	return nil, newNotFoundError(getNotFoundMessage("label", buildUCloudLabelID(key, value)))
}

func describeLabelAttachment(client *labelapi.LabelClient, key, value, resourceID string) (*labelapi.ListResourcesByLabelsResource, error) {
	limit := 100
	offset := 0
	listProjectsReq := client.NewListProjectsByLabelsRequest()
	listProjectsReq.Labels = []labelapi.ListProjectsByLabelsParamLabels{{Key: ucloud.String(key), Value: ucloud.String(value)}}
	listProjectsResp, err := client.ListProjectsByLabels(listProjectsReq)
	if err != nil {
		return nil, fmt.Errorf("error on listing projects by labels, %s", err)
	}
	projectIDs := make([]string, 0)
	resourceTypes := make([]string, 0)
	for _, project := range listProjectsResp.Projects {
		projectIDs = append(projectIDs, project.ProjectId)
		resourceTypes = append(resourceTypes, project.ResourceTypes...)
	}
	if len(projectIDs) > 0 && len(resourceTypes) > 0 {
		for {
			req := client.NewListResourcesByLabelsRequest()
			req.Limit = ucloud.Int(limit)
			req.Offset = ucloud.Int(offset)
			req.Labels = []labelapi.ListResourcesByLabelsParamLabels{{Key: ucloud.String(key), Value: ucloud.String(value)}}
			req.ProjectIds = projectIDs
			req.ResourceTypes = resourceTypes
			resp, err := client.ListResourcesByLabels(req)
			if err != nil {
				return nil, fmt.Errorf("error on listing resources by labels, %s", err)
			}
			for _, resourceInfo := range resp.Resources {
				if resourceInfo.ResourceId == resourceID {
					return &resourceInfo, nil
				}
			}
			if len(resp.Resources) < limit {
				break
			}
			offset += limit
		}
	}
	return nil, newNotFoundError(getNotFoundMessage("label attachment", buildUCloudLabelAttachmentID(key, value, resourceID)))
}
