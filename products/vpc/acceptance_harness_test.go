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

func isNotFoundCode(code int) bool {
	return code == 54002 || code == 58103
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
