package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"strconv"
	"strings"
)

func (client *UCloudClient) describeAccessKey(userName, accessKeyID string) (*iam.AccessKey, error) {
	if userName == "" {
		return nil, newNotFoundError(getNotFoundMessage("user_name", userName))
	}
	if accessKeyID == "" {
		return nil, newNotFoundError(getNotFoundMessage("access_key", accessKeyID))
	}
	req := client.iamconn.NewListAccessKeysRequest()
	req.UserName = ucloud.String(userName)

	resp, err := client.iamconn.ListAccessKeys(req)
	if err != nil {
		return nil, err
	}
	if len(resp.AccessKey) < 1 {
		return nil, newNotFoundError(getNotFoundMessage("access_key", accessKeyID))
	}

	for _, v := range resp.AccessKey {
		if accessKeyID == v.AccessKeyID {
			return &v, nil
		}
	}

	return nil, newNotFoundError(getNotFoundMessage("access_key", accessKeyID))
}

func (client *UCloudClient) describeAccessKeyByID(accessKeyID string) (*iam.AccessKey, error) {
	limit := 100
	offset := 0
	for {
		listUsersReq := client.iamconn.NewListUsersRequest()
		listUsersReq.Limit = ucloud.String(strconv.Itoa(limit))
		listUsersReq.Offset = ucloud.String(strconv.Itoa(offset))
		listUsersResp, err := client.iamconn.ListUsers(listUsersReq)
		if err != nil {
			return nil, err
		}
		if len(listUsersResp.Users) < 1 {
			break
		}
		for _, user := range listUsersResp.Users {
			accessKey, err := client.describeAccessKey(user.UserName, accessKeyID)
			if isNotFoundError(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			return accessKey, nil
		}
		if len(listUsersResp.Users) < limit {
			break
		}
		offset = offset + limit
	}
	return nil, newNotFoundError(getNotFoundMessage("access_key", accessKeyID))
}

func (client *UCloudClient) describeGroup(name string) (*iam.Group, error) {
	if name == "" {
		return nil, newNotFoundError(getNotFoundMessage("group", name))
	}
	req := client.iamconn.NewGetGroupRequest()
	req.GroupName = ucloud.String(name)

	resp, err := client.iamconn.GetGroup(req)
	if err != nil {
		if resp != nil && resp.RetCode == 11162 {
			return nil, newNotFoundError(getNotFoundMessage("group", name))
		}
		return nil, err
	}
	return &resp.Group, nil
}

func (client *UCloudClient) addUsersToGroup(users []string, group string) error {
	for _, u := range users {
		req := client.iamconn.NewAddUserToGroupRequest()
		req.GroupName = ucloud.String(group)
		req.UserName = ucloud.String(u)
		_, err := client.iamconn.AddUserToGroup(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *UCloudClient) removeUsersFromGroup(users []string, group string) error {
	for _, u := range users {
		req := client.iamconn.NewRemoveUserFromGroupRequest()
		req.GroupName = ucloud.String(group)
		req.UserName = ucloud.String(u)
		_, err := client.iamconn.RemoveUserFromGroup(req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *UCloudClient) describeGroupMembership(group string) ([]iam.UserForGroup, error) {
	limit := 100
	offset := 0
	users := make([]iam.UserForGroup, 0)
	for {
		req := client.iamconn.NewListUsersForGroupRequest()
		req.GroupName = ucloud.String(group)
		req.Limit = ucloud.String(strconv.Itoa(limit))
		req.Offset = ucloud.String(strconv.Itoa(offset))
		resp, err := client.iamconn.ListUsersForGroup(req)
		if err != nil {
			return nil, err
		}
		if len(resp.Users) < 1 {
			break
		}
		users = append(users, resp.Users...)
		if len(resp.Users) < limit {
			break
		}
		offset = offset + limit
	}

	return users, nil
}

func (client *UCloudClient) describeUser(name string) (*iam.User, error) {
	if name == "" {
		return nil, newNotFoundError(getNotFoundMessage("user", name))
	}
	req := client.iamconn.NewGetUserRequest()
	req.UserName = ucloud.String(name)

	resp, err := client.iamconn.GetUser(req)
	if err != nil {
		if resp != nil && resp.RetCode == 11021 {
			return nil, newNotFoundError(getNotFoundMessage("user", name))
		}
		return nil, err
	}
	//loginProfile
	return &resp.User, nil
}

func (client *UCloudClient) describeLoginProfile(name string) (*iam.LoginProfile, error) {
	if name == "" {
		return nil, newNotFoundError(getNotFoundMessage("login_profile", name))
	}
	req := client.iamconn.NewGetLoginProfileRequest()
	req.UserName = ucloud.String(name)

	resp, err := client.iamconn.GetLoginProfile(req)
	if err != nil {
		if resp != nil && resp.RetCode == 11021 {
			return nil, newNotFoundError(getNotFoundMessage("login_profile", name))
		}
		return nil, err
	}
	//loginProfile
	return &resp.LoginProfile, nil
}

func (client *UCloudClient) describeIAMProjectById(id string) (*iam.Project, error) {
	limit := 100
	for offset := 0; ; offset += limit {
		req := client.iamconn.NewListProjectsRequest()
		req.Limit = ucloud.String(strconv.Itoa(limit))
		req.Offset = ucloud.String(strconv.Itoa(offset))
		resp, err := client.iamconn.ListProjects(req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Projects) < 1 {
			return nil, newNotFoundError(getNotFoundMessage("iam", id))
		}
		for _, p := range resp.Projects {
			if p.ProjectID == id {
				return &p, nil
			}
		}
	}
}

func (client *UCloudClient) listIAMProject() ([]iam.Project, error) {
	limit := 100
	projects := make([]iam.Project, 0)
	for offset := 0; ; offset += limit {
		req := client.iamconn.NewListProjectsRequest()
		req.Limit = ucloud.String(strconv.Itoa(limit))
		req.Offset = ucloud.String(strconv.Itoa(offset))
		resp, err := client.iamconn.ListProjects(req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Projects) < 1 {
			return projects, nil
		}
		projects = append(projects, resp.Projects...)
	}
}

func (client *UCloudClient) describeIAMPolicyByName(name string, owner string) (*iam.IAMPolicy, error) {
	limit := 100
	for offset := 0; ; offset += limit {
		req := client.iamconn.NewListPoliciesRequest()
		req.Limit = ucloud.String(strconv.Itoa(limit))
		req.Offset = ucloud.String(strconv.Itoa(offset))
		req.Owner = ucloud.String(owner)
		resp, err := client.iamconn.ListPolicies(req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Policies) < 1 {
			return nil, newNotFoundError(getNotFoundMessage("iam", name))
		}
		for _, p := range resp.Policies {
			if p.PolicyName == name {
				return client.describeIAMPolicyByURN(p.PolicyURN)
			}
		}
	}
}

func (client *UCloudClient) describeIAMPolicyByURN(urn string) (*iam.IAMPolicy, error) {
	req := client.iamconn.NewGetIAMPolicyRequest()
	req.PolicyURN = ucloud.String(urn)
	resp, err := client.iamconn.GetIAMPolicy(req)
	if err != nil {
		if resp != nil && resp.RetCode == 11217 {
			return nil, newNotFoundError(getNotFoundMessage("iam", urn))
		}
		return nil, err
	}

	return &resp.Policy, nil
}

func (client *UCloudClient) describeIAMUserPolicyAttachment(userName string, policyURN string, projectID string) (*iam.Policy, error) {
	limit := 100
	for offset := 0; ; offset += limit {
		req := client.iamconn.NewListPoliciesForUserRequest()
		req.Limit = ucloud.String(strconv.Itoa(limit))
		req.Offset = ucloud.String(strconv.Itoa(offset))
		req.UserName = ucloud.String(userName)
		if projectID != "" {
			req.ProjectId = nil
			req.ProjectID = ucloud.String(projectID)
			req.Scope = ucloud.String("Specified")
		} else {
			req.Scope = ucloud.String("Unspecified")
		}

		resp, err := client.iamconn.ListPoliciesForUser(req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Policies) < 1 {
			return nil, newNotFoundError(getNotFoundMessage("iam user policy attachment", strings.Join([]string{userName, policyURN}, "/")))
		}
		for _, p := range resp.Policies {
			if p.PolicyURN == policyURN {
				return &p, nil
			}
		}
	}
}

func (client *UCloudClient) describeIAMGroupPolicyAttachment(groupName string, policyURN string, projectID string) (*iam.Policy, error) {
	limit := 100
	for offset := 0; ; offset += limit {
		req := client.iamconn.NewListPoliciesForGroupRequest()
		req.Limit = ucloud.String(strconv.Itoa(limit))
		req.Offset = ucloud.String(strconv.Itoa(offset))
		req.GroupName = ucloud.String(groupName)
		if projectID != "" {
			req.ProjectId = nil
			req.ProjectID = ucloud.String(projectID)
			req.Scope = ucloud.String("Specified")
		} else {
			req.Scope = ucloud.String("Unspecified")
		}

		resp, err := client.iamconn.ListPoliciesForGroup(req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Policies) < 1 {
			return nil, newNotFoundError(getNotFoundMessage("iam group policy attachment", strings.Join([]string{groupName, policyURN}, "/")))
		}
		for _, p := range resp.Policies {
			if p.PolicyURN == policyURN {
				return &p, nil
			}
		}
	}
}
