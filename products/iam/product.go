package iam

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

const Name = "iam"

type adapterV1 struct{}

var _ product.V1 = adapterV1{}

// New returns the stable product registration consumed by the provider.
func New() product.V1 {
	return adapterV1{}
}

func (adapterV1) Registration() product.Registration {
	return product.Registration{
		Name: Name,
		Resources: map[string]*schema.Resource{
			"ucloud_iam_access_key":              resourceUCloudIAMAccessKey(),
			"ucloud_iam_user":                    resourceUCloudIAMUser(),
			"ucloud_iam_group":                   resourceUCloudIAMGroup(),
			"ucloud_iam_group_membership":        resourceUCloudIAMGroupMembership(),
			"ucloud_iam_project":                 resourceUCloudIAMProject(),
			"ucloud_iam_policy":                  resourceUCloudIAMPolicy(),
			"ucloud_iam_user_policy_attachment":  resourceUCloudIAMUserPolicyAttachment(),
			"ucloud_iam_group_policy_attachment": resourceUCloudIAMGroupPolicyAttachment(),
		},
		DataSources: map[string]*schema.Resource{
			"ucloud_iam_users":           dataSourceUCloudIAMUsers(),
			"ucloud_iam_groups":          dataSourceUCloudIAMGroups(),
			"ucloud_iam_projects":        dataSourceUCloudIAMProjects(),
			"ucloud_iam_policy":          dataSourceUCloudIAMPolicy(),
			"ucloud_iam_policy_document": dataSourceUCloudIAMPolicyDocument(),
		},
	}
}
