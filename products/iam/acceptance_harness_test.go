package iam_test

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productiam "github.com/terraform-providers/terraform-provider-ucloud/products/iam"
	iamapi "github.com/ucloud/ucloud-sdk-go/services/iam"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

const (
	iamStatusInactive = "Inactive"

	accountScopeAttachmentPrefix = "account/"
	projectScopeAttachmentPrefix = "project/"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccIAMClient() (*iamapi.IAMClient, error) {
	client, err := testAccHarness.ProductClient(productiam.Name, newAccIAMClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*iamapi.IAMClient)
	if !ok {
		return nil, fmt.Errorf("unexpected IAM acceptance client type %T", client)
	}
	return typed, nil
}

func newAccIAMClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	client := iamapi.NewClient(config, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

func iamResponseStatus(err error, retCode int, message string) (bool, error) {
	if err != nil {
		if isIAMNotFoundError(err) {
			return true, nil
		}
		return false, err
	}
	if retCode == 0 {
		return false, nil
	}
	if isIAMNotFoundCode(retCode) {
		return true, nil
	}
	return false, fmt.Errorf("IAM API returned code %d: %s", retCode, message)
}

func isIAMNotFoundError(err error) bool {
	cloudErr, ok := err.(uerr.Error)
	return ok && isIAMNotFoundCode(cloudErr.Code())
}

func isIAMNotFoundCode(code int) bool {
	switch code {
	case 11021, 11162, 11217:
		return true
	default:
		return false
	}
}

func describeIAMAccessKey(client *iamapi.IAMClient, userName, accessKeyID string) (*iamapi.AccessKey, bool, error) {
	request := client.NewListAccessKeysRequest()
	request.UserName = ucloud.String(userName)
	response, err := client.ListAccessKeys(request)
	retCode, message := 0, ""
	if response != nil {
		retCode = response.GetRetCode()
		message = response.GetMessage()
	}
	notFound, err := iamResponseStatus(err, retCode, message)
	if err != nil || notFound {
		return nil, false, err
	}
	if response == nil || len(response.AccessKey) == 0 {
		return nil, false, nil
	}
	for index := range response.AccessKey {
		if response.AccessKey[index].AccessKeyID == accessKeyID {
			return &response.AccessKey[index], true, nil
		}
	}
	return nil, false, nil
}

func describeIAMAccessKeyByID(client *iamapi.IAMClient, accessKeyID string) (*iamapi.AccessKey, bool, error) {
	const limit = 100
	for offset := 0; ; offset += limit {
		request := client.NewListUsersRequest()
		request.Limit = ucloud.String(strconv.Itoa(limit))
		request.Offset = ucloud.String(strconv.Itoa(offset))
		response, err := client.ListUsers(request)
		retCode, message := 0, ""
		if response != nil {
			retCode = response.GetRetCode()
			message = response.GetMessage()
		}
		notFound, err := iamResponseStatus(err, retCode, message)
		if err != nil {
			return nil, false, err
		}
		if notFound || response == nil || len(response.Users) == 0 {
			return nil, false, nil
		}
		for _, user := range response.Users {
			accessKey, found, err := describeIAMAccessKey(client, user.UserName, accessKeyID)
			if err != nil {
				return nil, false, err
			}
			if found {
				return accessKey, true, nil
			}
		}
		if len(response.Users) < limit {
			return nil, false, nil
		}
	}
}

func describeIAMGroup(client *iamapi.IAMClient, name string) (*iamapi.Group, bool, error) {
	request := client.NewGetGroupRequest()
	request.GroupName = ucloud.String(name)
	response, err := client.GetGroup(request)
	retCode, message := 0, ""
	if response != nil {
		retCode = response.GetRetCode()
		message = response.GetMessage()
	}
	notFound, err := iamResponseStatus(err, retCode, message)
	if err != nil || notFound {
		return nil, false, err
	}
	if response == nil || response.Group.GroupName == "" {
		return nil, false, nil
	}
	return &response.Group, true, nil
}

func describeIAMGroupMembership(client *iamapi.IAMClient, group string) ([]iamapi.UserForGroup, bool, error) {
	const limit = 100
	users := make([]iamapi.UserForGroup, 0)
	for offset := 0; ; offset += limit {
		request := client.NewListUsersForGroupRequest()
		request.GroupName = ucloud.String(group)
		request.Limit = ucloud.String(strconv.Itoa(limit))
		request.Offset = ucloud.String(strconv.Itoa(offset))
		response, err := client.ListUsersForGroup(request)
		retCode, message := 0, ""
		if response != nil {
			retCode = response.GetRetCode()
			message = response.GetMessage()
		}
		notFound, err := iamResponseStatus(err, retCode, message)
		if err != nil {
			return nil, false, err
		}
		if notFound || response == nil {
			return nil, false, nil
		}
		users = append(users, response.Users...)
		if len(response.Users) < limit {
			return users, true, nil
		}
	}
}

func describeIAMUser(client *iamapi.IAMClient, name string) (*iamapi.User, bool, error) {
	request := client.NewGetUserRequest()
	request.UserName = ucloud.String(name)
	response, err := client.GetUser(request)
	retCode, message := 0, ""
	if response != nil {
		retCode = response.GetRetCode()
		message = response.GetMessage()
	}
	notFound, err := iamResponseStatus(err, retCode, message)
	if err != nil || notFound {
		return nil, false, err
	}
	if response == nil || response.User.UserName == "" {
		return nil, false, nil
	}
	return &response.User, true, nil
}

func describeIAMProjectByID(client *iamapi.IAMClient, id string) (*iamapi.Project, bool, error) {
	const limit = 100
	for offset := 0; ; offset += limit {
		request := client.NewListProjectsRequest()
		request.Limit = ucloud.String(strconv.Itoa(limit))
		request.Offset = ucloud.String(strconv.Itoa(offset))
		response, err := client.ListProjects(request)
		retCode, message := 0, ""
		if response != nil {
			retCode = response.GetRetCode()
			message = response.GetMessage()
		}
		notFound, err := iamResponseStatus(err, retCode, message)
		if err != nil {
			return nil, false, err
		}
		if notFound || response == nil || len(response.Projects) == 0 {
			return nil, false, nil
		}
		for index := range response.Projects {
			if response.Projects[index].ProjectID == id {
				return &response.Projects[index], true, nil
			}
		}
		if len(response.Projects) < limit {
			return nil, false, nil
		}
	}
}

func describeIAMPolicyByName(client *iamapi.IAMClient, name, owner string) (*iamapi.IAMPolicy, bool, error) {
	const limit = 100
	for offset := 0; ; offset += limit {
		request := client.NewListPoliciesRequest()
		request.Limit = ucloud.String(strconv.Itoa(limit))
		request.Offset = ucloud.String(strconv.Itoa(offset))
		request.Owner = ucloud.String(owner)
		response, err := client.ListPolicies(request)
		retCode, message := 0, ""
		if response != nil {
			retCode = response.GetRetCode()
			message = response.GetMessage()
		}
		notFound, err := iamResponseStatus(err, retCode, message)
		if err != nil {
			return nil, false, err
		}
		if notFound || response == nil || len(response.Policies) == 0 {
			return nil, false, nil
		}
		for _, policy := range response.Policies {
			if policy.PolicyName == name {
				return describeIAMPolicyByURN(client, policy.PolicyURN)
			}
		}
		if len(response.Policies) < limit {
			return nil, false, nil
		}
	}
}

func describeIAMPolicyByURN(client *iamapi.IAMClient, urn string) (*iamapi.IAMPolicy, bool, error) {
	request := client.NewGetIAMPolicyRequest()
	request.PolicyURN = ucloud.String(urn)
	response, err := client.GetIAMPolicy(request)
	retCode, message := 0, ""
	if response != nil {
		retCode = response.GetRetCode()
		message = response.GetMessage()
	}
	notFound, err := iamResponseStatus(err, retCode, message)
	if err != nil || notFound {
		return nil, false, err
	}
	if response == nil || response.Policy.PolicyURN == "" {
		return nil, false, nil
	}
	return &response.Policy, true, nil
}

func describeIAMUserPolicyAttachment(client *iamapi.IAMClient, userName, policyURN, projectID string) (*iamapi.Policy, bool, error) {
	const limit = 100
	for offset := 0; ; offset += limit {
		request := client.NewListPoliciesForUserRequest()
		request.Limit = ucloud.String(strconv.Itoa(limit))
		request.Offset = ucloud.String(strconv.Itoa(offset))
		request.UserName = ucloud.String(userName)
		if projectID == "" {
			request.Scope = ucloud.String("Unspecified")
		} else {
			request.ProjectId = nil
			request.ProjectID = ucloud.String(projectID)
			request.Scope = ucloud.String("Specified")
		}
		response, err := client.ListPoliciesForUser(request)
		retCode, message := 0, ""
		if response != nil {
			retCode = response.GetRetCode()
			message = response.GetMessage()
		}
		notFound, err := iamResponseStatus(err, retCode, message)
		if err != nil {
			return nil, false, err
		}
		if notFound || response == nil || len(response.Policies) == 0 {
			return nil, false, nil
		}
		for index := range response.Policies {
			if response.Policies[index].PolicyURN == policyURN {
				return &response.Policies[index], true, nil
			}
		}
		if len(response.Policies) < limit {
			return nil, false, nil
		}
	}
}

func describeIAMGroupPolicyAttachment(client *iamapi.IAMClient, groupName, policyURN, projectID string) (*iamapi.Policy, bool, error) {
	const limit = 100
	for offset := 0; ; offset += limit {
		request := client.NewListPoliciesForGroupRequest()
		request.Limit = ucloud.String(strconv.Itoa(limit))
		request.Offset = ucloud.String(strconv.Itoa(offset))
		request.GroupName = ucloud.String(groupName)
		if projectID == "" {
			request.Scope = ucloud.String("Unspecified")
		} else {
			request.ProjectId = nil
			request.ProjectID = ucloud.String(projectID)
			request.Scope = ucloud.String("Specified")
		}
		response, err := client.ListPoliciesForGroup(request)
		retCode, message := 0, ""
		if response != nil {
			retCode = response.GetRetCode()
			message = response.GetMessage()
		}
		notFound, err := iamResponseStatus(err, retCode, message)
		if err != nil {
			return nil, false, err
		}
		if notFound || response == nil || len(response.Policies) == 0 {
			return nil, false, nil
		}
		for index := range response.Policies {
			if response.Policies[index].PolicyURN == policyURN {
				return &response.Policies[index], true, nil
			}
		}
		if len(response.Policies) < limit {
			return nil, false, nil
		}
	}
}

func testAccCheckIAMAccessKeyExists(name string, value *iamapi.AccessKey) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("vpc id is empty")
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		accessKey, found, err := describeIAMAccessKey(client, item.Primary.Attributes["user_name"], item.Primary.ID)
		log.Printf("[INFO] access key id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("access key %q was not found", item.Primary.ID)
		}
		*value = *accessKey
		return nil
	}
}

func testAccCheckIAMAccessKeyAttributes(value *iamapi.AccessKey) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.AccessKeyID == "" {
			return fmt.Errorf("access key id is empty")
		}
		return nil
	}
}

func testAccCheckIAMAccessKeyDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_access_key" {
			continue
		}
		_, found, err := describeIAMAccessKeyByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if found {
			return fmt.Errorf("IAM access key still exist")
		}
	}
	return nil
}

func testAccCheckIAMGroupExists(name string, value *iamapi.Group) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("group name is empty")
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		group, found, err := describeIAMGroup(client, item.Primary.ID)
		log.Printf("[INFO] group name %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("group %q was not found", item.Primary.ID)
		}
		*value = *group
		return nil
	}
}

func testAccCheckIAMGroupAttributes(value *iamapi.Group) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.GroupName == "" {
			return fmt.Errorf("group name is empty")
		}
		return nil
	}
}

func testAccCheckIAMGroupDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_group" {
			continue
		}
		group, found, err := describeIAMGroup(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && group.GroupName != "" {
			return fmt.Errorf("group still exist")
		}
	}
	return nil
}

func testAccCheckIAMGroupMembershipExists(name string, users *[]iamapi.UserForGroup) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("id is empty")
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		value, found, err := describeIAMGroupMembership(client, item.Primary.ID)
		log.Printf("[INFO] group name %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("group membership %q was not found", item.Primary.ID)
		}
		*users = value
		return nil
	}
}

func testAccCheckIAMGroupMembershipAttributes(users *[]iamapi.UserForGroup, size int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if len(*users) != size {
			return fmt.Errorf("length of val is not %v", size)
		}
		return nil
	}
}

func testAccCheckIAMPolicyExists(name string, value *iamapi.IAMPolicy) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("policy name is empty")
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		policy, found, err := describeIAMPolicyByName(client, item.Primary.ID, "User")
		log.Printf("[INFO] policy name %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("policy %q was not found", item.Primary.ID)
		}
		*value = *policy
		return nil
	}
}

func testAccCheckIAMPolicyAttributes(value *iamapi.IAMPolicy) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.PolicyName == "" {
			return fmt.Errorf("policy name is empty")
		}
		return nil
	}
}

func testAccCheckIAMPolicyDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_policy" {
			continue
		}
		policy, found, err := describeIAMPolicyByName(client, item.Primary.ID, "User")
		if err != nil {
			return err
		}
		if found && policy.PolicyName != "" {
			return fmt.Errorf("policy still exist")
		}
	}
	return nil
}

func testAccCheckIAMProjectExists(name string, value *iamapi.Project) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("project id is empty")
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		project, found, err := describeIAMProjectByID(client, item.Primary.ID)
		log.Printf("[INFO] project id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("project %q was not found", item.Primary.ID)
		}
		*value = *project
		return nil
	}
}

func testAccCheckIAMProjectAttributes(value *iamapi.Project) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.ProjectName == "" {
			return fmt.Errorf("project name is empty")
		}
		return nil
	}
}

func testAccCheckIAMProjectDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_project" {
			continue
		}
		project, found, err := describeIAMProjectByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && project.ProjectName != "" {
			return fmt.Errorf("project still exist")
		}
	}
	return nil
}

func testAccCheckIAMUserExists(name string, value *iamapi.User) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("user name is empty")
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		user, found, err := describeIAMUser(client, item.Primary.ID)
		log.Printf("[INFO] user name id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("user %q was not found", item.Primary.ID)
		}
		*value = *user
		return nil
	}
}

func testAccCheckIAMUserAttributes(value *iamapi.User) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.UserName == "" {
			return fmt.Errorf("user name is empty")
		}
		return nil
	}
}

func testAccCheckIAMUserDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_user" {
			continue
		}
		user, found, err := describeIAMUser(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && user.UserName != "" {
			return fmt.Errorf("user still exist")
		}
	}
	return nil
}

func testAccCheckIAMUserPolicyAttachmentExists(name string, value *iamapi.Policy) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("user name is empty")
		}
		userName, policyURN, projectID, err := extractIAMPolicyAttachmentID(item.Primary.ID)
		if err != nil {
			return err
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		policy, found, err := describeIAMUserPolicyAttachment(client, userName, policyURN, projectID)
		log.Printf("[INFO] user name id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("user policy attachment %q was not found", item.Primary.ID)
		}
		*value = *policy
		return nil
	}
}

func testAccCheckIAMUserPolicyAttachmentAttributes(value *iamapi.Policy) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.PolicyName == "" {
			return fmt.Errorf("policy name is empty")
		}
		return nil
	}
}

func testAccCheckIAMUserPolicyAttachmentDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_user" {
			continue
		}
		user, found, err := describeIAMUser(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && user.UserName != "" {
			return fmt.Errorf("user still exist")
		}
	}
	return nil
}

func testAccCheckIAMGroupPolicyAttachmentExists(name string, value *iamapi.Policy) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("group policy attachment is empty")
		}
		groupName, policyURN, projectID, err := extractIAMPolicyAttachmentID(item.Primary.ID)
		if err != nil {
			return err
		}
		client, err := testAccIAMClient()
		if err != nil {
			return err
		}
		policy, found, err := describeIAMGroupPolicyAttachment(client, groupName, policyURN, projectID)
		log.Printf("[INFO] group policy attachment id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("group policy attachment %q was not found", item.Primary.ID)
		}
		*value = *policy
		return nil
	}
}

func testAccCheckIAMGroupPolicyAttachmentAttributes(value *iamapi.Policy) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.PolicyName == "" {
			return fmt.Errorf("policy name is empty")
		}
		return nil
	}
}

func testAccCheckIAMGroupPolicyAttachmentDestroy(state *terraform.State) error {
	client, err := testAccIAMClient()
	if err != nil {
		return err
	}
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_iam_group" {
			continue
		}
		group, found, err := describeIAMGroup(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && group.GroupName != "" {
			return fmt.Errorf("group still exist")
		}
	}
	return nil
}

func extractIAMPolicyAttachmentID(id string) (entityName, policyURN, projectID string, err error) {
	if strings.HasPrefix(id, accountScopeAttachmentPrefix) {
		items := strings.SplitN(strings.TrimPrefix(id, accountScopeAttachmentPrefix), "/", 2)
		if len(items) != 2 || items[0] == "" || items[1] == "" {
			return "", "", "", fmt.Errorf("fail to parse id")
		}
		return items[0], items[1], "", nil
	}
	if strings.HasPrefix(id, projectScopeAttachmentPrefix) {
		items := strings.SplitN(strings.TrimPrefix(id, projectScopeAttachmentPrefix), "/", 3)
		if len(items) != 3 || items[0] == "" || items[1] == "" || items[2] == "" {
			return "", "", "", fmt.Errorf("fail to parse id")
		}
		return items[1], items[2], items[0], nil
	}
	return "", "", "", fmt.Errorf("fail to parse id")
}
