package vpc_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productvpc "github.com/terraform-providers/terraform-provider-ucloud/products/vpc"
	unetapi "github.com/ucloud/ucloud-sdk-go/services/unet"
	vpcapi "github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

const defaultTag = "Default"

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

type acceptanceClients struct {
	vpcconn  *vpcapi.VPCClient
	unetconn *unetapi.UNetClient
	region   string
}

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccClients() (*acceptanceClients, error) {
	client, err := testAccHarness.ProductClient(productvpc.Name, func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		clients := &acceptanceClients{
			vpcconn:  vpcapi.NewClient(config, credential),
			unetconn: unetapi.NewClient(config, credential),
			region:   config.Region,
		}
		for _, handler := range handlers {
			_ = clients.vpcconn.AddHttpRequestHandler(handler)
			_ = clients.unetconn.AddHttpRequestHandler(handler)
		}
		return clients
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*acceptanceClients)
	if !ok {
		return nil, fmt.Errorf("unexpected VPC acceptance client type %T", client)
	}
	return typed, nil
}

// isNotFoundCode lists the retcodes this product treats as "the resource is
// gone". 54002 is the generic one, 58103 is the VPC backend's, and 208704 is
// the security group backend's - the one a destroyed group is reported with,
// which is what the sec group destroy check has to recognize.
func isNotFoundCode(code int) bool {
	return code == 54002 || code == 58103 || code == 208704
}

func isNotFoundError(err error) bool {
	cloudErr, ok := err.(uerr.Error)
	return ok && isNotFoundCode(cloudErr.Code())
}

func describeAccVPCByID(client *vpcapi.VPCClient, id string) (*vpcapi.VPCInfo, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("vpc id is empty")
	}
	request := client.NewDescribeVPCRequest()
	request.VPCIds = []string{id}
	response, err := client.DescribeVPC(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading vpc %q, %s", id, response.GetMessage())
	}
	if len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccSubnetByID(client *vpcapi.VPCClient, id string) (*vpcapi.SubnetInfo, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("subnet id is empty")
	}
	request := client.NewDescribeSubnetRequest()
	request.SubnetIds = []string{id}
	response, err := client.DescribeSubnet(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading subnet %q, %s", id, response.GetMessage())
	}
	if len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccVIPByID(client *vpcapi.VPCClient, id string) (*vpcapi.VIPDetailSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("vip id is empty")
	}
	request := client.NewDescribeVIPRequest()
	request.VIPId = ucloud.String(id)
	response, err := client.DescribeVIP(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading vip %q, %s", id, response.GetMessage())
	}
	if len(response.VIPSet) == 0 {
		return nil, false, nil
	}
	return &response.VIPSet[0], true, nil
}

func describeAccNatGatewayByID(client *vpcapi.VPCClient, id string) (*vpcapi.NatGatewayDataSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("nat gateway id is empty")
	}
	request := client.NewDescribeNATGWRequest()
	request.NATGWIds = []string{id}
	response, err := client.DescribeNATGW(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading nat gateway %q, %s", id, response.GetMessage())
	}
	if len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccNatGatewayRuleByID(client *vpcapi.VPCClient, policyID, natGatewayID string) (*vpcapi.NATGWPolicyDataSet, bool, error) {
	if policyID == "" {
		return nil, false, fmt.Errorf("nat gateway rule id is empty")
	}
	request := client.NewDescribeNATGWPolicyRequest()
	request.NATGWId = ucloud.String(natGatewayID)
	response, err := client.DescribeNATGWPolicy(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading nat gateway rule %q, %s", policyID, response.GetMessage())
	}
	for index := range response.DataSet {
		if response.DataSet[index].PolicyId == policyID {
			return &response.DataSet[index], true, nil
		}
	}
	return nil, false, nil
}

func describeAccSecGroupByID(client *vpcapi.VPCClient, id string) (*vpcapi.SecGroupInfo, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("sec group id is empty")
	}
	request := client.NewDescribeSecGroupRequest()
	request.SecGroupId = []string{id}
	response, err := client.DescribeSecGroup(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading sec group %q, %s", id, response.GetMessage())
	}
	if len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func describeAccEIPByID(client *unetapi.UNetClient, id string) (*unetapi.UnetEIPSet, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("eip id is empty")
	}
	request := client.NewDescribeEIPRequest()
	request.EIPIds = []string{id}
	response, err := client.DescribeEIP(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading eip %q, %s", id, response.GetMessage())
	}
	if len(response.EIPSet) == 0 {
		return nil, false, nil
	}
	return &response.EIPSet[0], true, nil
}

func testAccCheckVPCExists(name string, target *vpcapi.VPCInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("vpc id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccVPCByID(clients.vpcconn, item.Primary.ID)
		log.Printf("[INFO] vpc id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("vpc %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckVPCAttributes(value *vpcapi.VPCInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.VPCId == "" {
			return fmt.Errorf("vpc id is empty")
		}
		return nil
	}
}

func testAccCheckVPCDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_vpc" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccVPCByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.VPCId != "" {
			return fmt.Errorf("VPC still exist")
		}
	}
	return nil
}

func testAccCheckSubnetExists(name string, target *vpcapi.SubnetInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("subnet id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSubnetByID(clients.vpcconn, item.Primary.ID)
		log.Printf("[INFO] subnet id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("subnet %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckSubnetAttributes(value *vpcapi.SubnetInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.SubnetId == "" {
			return fmt.Errorf("subnet id is empty")
		}
		if value.VPCId == "" {
			return fmt.Errorf("vpc id has not been bound")
		}
		return nil
	}
}

func testAccCheckSubnetDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_subnet" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSubnetByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.SubnetId != "" {
			return fmt.Errorf("subnet still exist")
		}
	}
	return nil
}

func testAccCheckVIPExists(name string, target *vpcapi.VIPDetailSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("vip id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccVIPByID(clients.vpcconn, item.Primary.ID)
		log.Printf("[INFO] vip id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("vip %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckVIPAttributes(value *vpcapi.VIPDetailSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.VIPId == "" {
			return fmt.Errorf("vip id is empty")
		}
		return nil
	}
}

func testAccCheckVIPDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_vip" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccVIPByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.VIPId != "" {
			return fmt.Errorf("vip still exist")
		}
	}
	return nil
}

func testAccCheckSecGroupExists(name string, target *vpcapi.SecGroupInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("sec group id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSecGroupByID(clients.vpcconn, item.Primary.ID)
		log.Printf("[INFO] sec group id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("sec group %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

// testAccCheckSecGroupAttributes asserts what the API holds for a group. Rules
// are deliberately not part of it: they belong to ucloud_sec_group_rule now, so
// a group is a valid group whether or not anything has been added to it.
func testAccCheckSecGroupAttributes(value *vpcapi.SecGroupInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.SecGroupId == "" {
			return fmt.Errorf("sec group id is empty")
		}
		if value.VPCId == "" {
			return fmt.Errorf("sec group has not been bound to a VPC")
		}
		return nil
	}
}

func testAccCheckSecGroupDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_sec_group" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSecGroupByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.SecGroupId != "" {
			return fmt.Errorf("sec group still exist")
		}
	}
	return nil
}

// describeAccSecGroupRuleByID looks a rule up inside its group. DescribeSecGroup
// has no RuleId filter, so the group is fetched and its rules searched.
func describeAccSecGroupRuleByID(client *vpcapi.VPCClient, secGroupID, ruleID string) (*vpcapi.SecGroupRuleInfo, bool, error) {
	if secGroupID == "" || ruleID == "" {
		return nil, false, fmt.Errorf("sec group rule needs both a group id and a rule id")
	}
	secGroup, found, err := describeAccSecGroupByID(client, secGroupID)
	if err != nil || !found {
		return nil, false, err
	}
	for index := range secGroup.Rule {
		if secGroup.Rule[index].RuleId == ruleID {
			return &secGroup.Rule[index], true, nil
		}
	}
	return nil, false, nil
}

func testAccCheckSecGroupRuleExists(name string, target *vpcapi.SecGroupRuleInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("sec group rule id is empty")
		}
		secGroupID := item.Primary.Attributes["sec_group_id"]
		if secGroupID == "" {
			return fmt.Errorf("sec group rule %q has no sec_group_id", item.Primary.ID)
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSecGroupRuleByID(clients.vpcconn, secGroupID, item.Primary.ID)
		log.Printf("[INFO] sec group rule id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("sec group rule %q is not found in sec group %q", item.Primary.ID, secGroupID)
		}
		*target = *value
		return nil
	}
}

// testAccCheckSecGroupRuleAttributes asserts what the API holds for a rule. The
// ip range is compared whole: a rule carrying several CIDR blocks stays one
// rule, so the comma separated string has to come back exactly as it went in.
func testAccCheckSecGroupRuleAttributes(value *vpcapi.SecGroupRuleInfo, wantIPRange string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.RuleId == "" {
			return fmt.Errorf("sec group rule has no id")
		}
		if value.IPRange != wantIPRange {
			return fmt.Errorf("sec group rule ip range = %q, want %q", value.IPRange, wantIPRange)
		}
		return nil
	}
}

// testAccCheckSecGroupRuleCount asserts how many rules the group actually holds.
// It is what catches a rule carrying several CIDR blocks being fanned out into
// several rules, which would break the one rule one resource model.
func testAccCheckSecGroupRuleCount(secGroupName string, want int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[secGroupName]
		if !ok {
			return fmt.Errorf("not found: %s", secGroupName)
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSecGroupByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("sec group %q is not found", item.Primary.ID)
		}
		if len(value.Rule) != want {
			return fmt.Errorf("sec group %q holds %d rules, want %d", item.Primary.ID, len(value.Rule), want)
		}
		return nil
	}
}

func testAccCheckSecGroupRuleDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_sec_group_rule" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSecGroupRuleByID(clients.vpcconn, item.Primary.Attributes["sec_group_id"], item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.RuleId != "" {
			return fmt.Errorf("sec group rule still exist")
		}
	}
	return nil
}

func testAccCheckNatGWExists(name string, target *vpcapi.NatGatewayDataSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("nat gateway id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccNatGatewayByID(clients.vpcconn, item.Primary.ID)
		log.Printf("[INFO] nat gateway id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("nat gateway %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckNatGWAttributes(value *vpcapi.NatGatewayDataSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.NATGWId == "" {
			return fmt.Errorf("nat gateway id is empty")
		}
		return nil
	}
}

func testAccCheckNatGWDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_nat_gateway" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccNatGatewayByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.NATGWId != "" {
			return fmt.Errorf("nat gateway still exist")
		}
	}
	return nil
}

func testAccCheckNatGWRuleExists(name string, natGateway *vpcapi.NatGatewayDataSet, target *vpcapi.NATGWPolicyDataSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("nat_gateway rule id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccNatGatewayRuleByID(clients.vpcconn, item.Primary.ID, natGateway.NATGWId)
		log.Printf("[INFO] nat_gateway rule id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("nat gateway rule %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckNatGWRuleAttributes(value *vpcapi.NATGWPolicyDataSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.PolicyId == "" {
			return fmt.Errorf("nat_gateway rule id is empty")
		}
		return nil
	}
}

func testAccCheckNatGWRuleDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_nat_gateway_rule" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccNatGatewayRuleByID(
			clients.vpcconn,
			item.Primary.ID,
			item.Primary.Attributes["nat_gateway_id"],
		)
		if err != nil {
			return err
		}
		if found && value.PolicyId != "" {
			return fmt.Errorf("nat_gateway rule still exist")
		}
	}
	return nil
}

func testAccCheckEIPExists(name string, target *unetapi.UnetEIPSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("eip id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccEIPByID(clients.unetconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("eip %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func describeAccRouteTableByID(client *vpcapi.VPCClient, id string) (*vpcapi.RouteTableInfo, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("route table id is empty")
	}
	request := client.NewDescribeRouteTableRequest()
	request.RouteTableId = ucloud.String(id)
	response, err := client.DescribeRouteTable(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading route table %q, %s", id, response.GetMessage())
	}
	if len(response.RouteTables) == 0 {
		return nil, false, nil
	}
	return &response.RouteTables[0], true, nil
}

func testAccCheckRouteTableExists(name string, target *vpcapi.RouteTableInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("route table id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccRouteTableByID(clients.vpcconn, item.Primary.ID)
		log.Printf("[INFO] route table id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("route table %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckRouteTableAttributes(value *vpcapi.RouteTableInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.RouteTableId == "" {
			return fmt.Errorf("route table id is empty")
		}
		if value.VPCId == "" {
			return fmt.Errorf("route table vpc id is empty")
		}
		return nil
	}
}

func testAccCheckRouteTableDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_route_table" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccRouteTableByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.RouteTableId != "" {
			return fmt.Errorf("route table still exist")
		}
	}
	return nil
}

func describeAccRouteRuleByID(client *vpcapi.VPCClient, routeTableID, routeRuleID string) (*vpcapi.RouteRuleInfo, bool, error) {
	if routeTableID == "" || routeRuleID == "" {
		return nil, false, fmt.Errorf("route table id or route rule id is empty")
	}
	request := client.NewDescribeRouteTableRequest()
	request.RouteTableId = ucloud.String(routeTableID)
	response, err := client.DescribeRouteTable(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading route table %q, %s", routeTableID, response.GetMessage())
	}
	if len(response.RouteTables) == 0 {
		return nil, false, nil
	}
	for i := range response.RouteTables[0].RouteRules {
		if response.RouteTables[0].RouteRules[i].RouteRuleId == routeRuleID {
			return &response.RouteTables[0].RouteRules[i], true, nil
		}
	}
	return nil, false, nil
}

func testAccCheckRouteTableRuleExists(name string, target *vpcapi.RouteRuleInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("route table rule id is empty")
		}
		routeTableID := item.Primary.Attributes["route_table_id"]
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccRouteRuleByID(clients.vpcconn, routeTableID, item.Primary.ID)
		log.Printf("[INFO] route table rule id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("route table rule %q is not found", item.Primary.ID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckRouteTableRuleAttributes(value *vpcapi.RouteRuleInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.RouteRuleId == "" {
			return fmt.Errorf("route table rule id is empty")
		}
		if value.DstAddr == "" {
			return fmt.Errorf("route table rule dst addr is empty")
		}
		if value.NexthopId == "" {
			return fmt.Errorf("route table rule nexthop id is empty")
		}
		return nil
	}
}

func testAccCheckRouteTableRuleDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_route_table_rule" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		routeTableID := item.Primary.Attributes["route_table_id"]
		value, found, err := describeAccRouteRuleByID(clients.vpcconn, routeTableID, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.RouteRuleId != "" {
			return fmt.Errorf("route table rule still exist")
		}
	}
	return nil
}

func testAccCheckRouteTableAssociationExists(name string, target *vpcapi.SubnetInfo) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		subnetID := item.Primary.Attributes["subnet_id"]
		if subnetID == "" {
			return fmt.Errorf("route table association subnet id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccSubnetByID(clients.vpcconn, subnetID)
		log.Printf("[INFO] route table association subnet id %#v", subnetID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("subnet %q for route table association is not found", subnetID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckRouteTableAssociationAttributes(value *vpcapi.SubnetInfo) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.SubnetId == "" {
			return fmt.Errorf("route table association subnet id is empty")
		}
		if value.RouteTableId == "" {
			return fmt.Errorf("route table association route table id is empty")
		}
		return nil
	}
}

func testAccCheckRouteTableAssociationDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_route_table_association" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		subnetID := item.Primary.Attributes["subnet_id"]
		routeTableID := item.Primary.Attributes["route_table_id"]
		value, found, err := describeAccSubnetByID(clients.vpcconn, subnetID)
		if err != nil {
			return err
		}
		if found && value.RouteTableId == routeTableID {
			return fmt.Errorf("route table association still exist")
		}
	}
	return nil
}

func describeAccNetworkInterfaceByID(client *vpcapi.VPCClient, id string) (*vpcapi.NetworkInterface, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("network interface id is empty")
	}
	request := client.NewDescribeNetworkInterfaceRequest()
	request.InterfaceId = []string{id}
	response, err := client.DescribeNetworkInterface(request)
	if err != nil {
		if isNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil {
		return nil, false, nil
	}
	if response.GetRetCode() != 0 {
		if isNotFoundCode(response.GetRetCode()) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error on reading network interface %q, %s", id, response.GetMessage())
	}
	if len(response.NetworkInterfaceSet) == 0 {
		return nil, false, nil
	}
	return &response.NetworkInterfaceSet[0], true, nil
}

func testAccCheckNetworkInterfaceExists(name string, target *vpcapi.NetworkInterface) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		interfaceID := item.Primary.ID
		if interfaceID == "" {
			return fmt.Errorf("network interface id is empty")
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccNetworkInterfaceByID(clients.vpcconn, interfaceID)
		log.Printf("[INFO] network interface id %#v", interfaceID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("network interface %q is not found", interfaceID)
		}
		*target = *value
		return nil
	}
}

func testAccCheckNetworkInterfaceAttributes(value *vpcapi.NetworkInterface) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if value.InterfaceId == "" {
			return fmt.Errorf("network interface id is empty")
		}
		if value.SubnetId == "" {
			return fmt.Errorf("network interface subnet id is empty")
		}
		if value.VPCId == "" {
			return fmt.Errorf("network interface vpc id is empty")
		}
		return nil
	}
}

func testAccCheckNetworkInterfaceDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_network_interface" {
			continue
		}
		clients, err := testAccClients()
		if err != nil {
			return err
		}
		value, found, err := describeAccNetworkInterfaceByID(clients.vpcconn, item.Primary.ID)
		if err != nil {
			return err
		}
		if found && value.InterfaceId != "" {
			return fmt.Errorf("network interface still exist")
		}
	}
	return nil
}
