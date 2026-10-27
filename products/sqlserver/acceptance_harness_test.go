package sqlserver_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productsqlserver "github.com/terraform-providers/terraform-provider-ucloud/products/sqlserver"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func newAccSQLServerClient(
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

func testAccSQLServerClient() (*ucloud.Client, error) {
	client, err := testAccHarness.ProductClient(productsqlserver.Name, newAccSQLServerClient)
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

func accDescribeSQLServerInstance(client *ucloud.Client, dbId, zone string) (map[string]interface{}, error) {
	cfg := client.GetConfig()
	payload := map[string]interface{}{
		"Region":    cfg.Region,
		"ProjectId": cfg.ProjectId,
		"Zone":      zone,
		"ClassType": "sqlserver",
		"DBId":      dbId,
	}
	resp, err := accInvoke(client, "DescribeUDBInstance", payload)
	if err != nil {
		return nil, err
	}
	dataSet, _ := resp["DataSet"].([]interface{})
	if len(dataSet) == 0 {
		return nil, fmt.Errorf("sqlserver instance %q not found", dbId)
	}
	item, ok := dataSet[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected describe item type %T", dataSet[0])
	}
	return item, nil
}

func testAccCheckSQLServerInstanceExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		client, err := testAccSQLServerClient()
		if err != nil {
			return err
		}
		_, err = accDescribeSQLServerInstance(client, item.Primary.ID, item.Primary.Attributes["availability_zone"])
		return err
	}
}

func testAccCheckSQLServerInstanceDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_sqlserver_instance" {
			continue
		}
		client, err := testAccSQLServerClient()
		if err != nil {
			return err
		}
		if _, err := accDescribeSQLServerInstance(client, item.Primary.ID, item.Primary.Attributes["availability_zone"]); err == nil {
			return fmt.Errorf("sqlserver instance %q still exists", item.Primary.ID)
		}
	}
	return nil
}
