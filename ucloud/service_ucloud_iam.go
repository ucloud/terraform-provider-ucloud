package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"strconv"
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
	if len(users) == 0 {
		return nil, newNotFoundError(getNotFoundMessage("group_membership", group))
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
