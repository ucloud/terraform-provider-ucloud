package udb_mysql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productmysql "github.com/terraform-providers/terraform-provider-ucloud/products/udb_mysql"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func newAccMySQLClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	cfg := *config
	cfg.Timeout = 60 * time.Second
	client := ucloud.NewClient(&cfg, credential)
	for _, handler := range handlers {
		_ = client.AddHttpRequestHandler(handler)
	}
	return client
}

func testAccMySQLClient() (*ucloud.Client, error) {
	client, err := testAccHarness.ProductClient(productmysql.Name, newAccMySQLClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*ucloud.Client)
	if !ok {
		return nil, fmt.Errorf("unexpected acceptance client type %T", client)
	}
	return typed, nil
}

func accInvoke(client *ucloud.Client, action string, payload map[string]interface{}) (map[string]interface{}, error) {
	payload["Action"] = action
	req := client.NewGenericRequest()
	if err := req.SetPayload(payload); err != nil {
		return nil, err
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	if resp.GetRetCode() != 0 {
		return nil, fmt.Errorf("api %s ret code %d: %s", action, resp.GetRetCode(), resp.GetMessage())
	}
	return resp.GetPayload(), nil
}

func accDescribeMySQLInstance(client *ucloud.Client, dbId, zone string) (map[string]interface{}, error) {
	cfg := client.GetConfig()
	payload := map[string]interface{}{
		"Region":    cfg.Region,
		"ProjectId": cfg.ProjectId,
		"Zone":      zone,
		"ClassType": "sql",
		"DBId":      dbId,
	}
	resp, err := accInvoke(client, "DescribeUDBInstance", payload)
	if err != nil {
		return nil, err
	}
	dataSet, _ := resp["DataSet"].([]interface{})
	if len(dataSet) == 0 {
		return nil, fmt.Errorf("mysql instance %q not found", dbId)
	}
	item, ok := dataSet[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected describe item type %T", dataSet[0])
	}
	return item, nil
}

func testAccCheckMySQLInstanceExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		client, err := testAccMySQLClient()
		if err != nil {
			return err
		}
		_, err = accDescribeMySQLInstance(client, item.Primary.ID, item.Primary.Attributes["availability_zone"])
		return err
	}
}

func testAccCheckMySQLInstanceDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_udb_mysql_instance" {
			continue
		}
		client, err := testAccMySQLClient()
		if err != nil {
			return err
		}
		if _, err := accDescribeMySQLInstance(client, item.Primary.ID, item.Primary.Attributes["availability_zone"]); err == nil {
			return fmt.Errorf("mysql instance %q still exists", item.Primary.ID)
		}
	}
	return nil
}
