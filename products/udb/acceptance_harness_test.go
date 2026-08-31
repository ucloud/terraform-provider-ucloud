package udb_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	udbapi "github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/acceptancetest"
	productudb "github.com/terraform-providers/terraform-provider-ucloud/products/udb"
)

var testAccHarness = acceptancetest.New()

var testAccProviders = testAccHarness.Providers

func testAccPreCheck(t *testing.T) {
	testAccHarness.PreCheck(t)
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return acceptancetest.CheckIDExists(name)
}

func testAccUDBClient() (*udbapi.UDBClient, error) {
	client, err := testAccHarness.ProductClient(productudb.Name, newAccUDBClient)
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*udbapi.UDBClient)
	if !ok {
		return nil, fmt.Errorf("unexpected UDB acceptance client type %T", client)
	}
	return typed, nil
}

func newAccUDBClient(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{} {
	longTimeoutConfig := *config
	longTimeoutConfig.Timeout = 60 * time.Second
	client := udbapi.NewClient(&longTimeoutConfig, credential)
	for _, handler := range handlers {
		client.AddHttpRequestHandler(handler)
	}
	return client
}

func describeAccDBInstanceByID(client *udbapi.UDBClient, id string) (*udbapi.UDBInstanceSet, bool, error) {
	if id == "" {
		return nil, false, nil
	}
	request := client.NewDescribeUDBInstanceRequest()
	request.DBId = ucloud.String(id)
	response, err := client.DescribeUDBInstance(request)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 230 {
			return nil, false, nil
		}
		return nil, false, err
	}
	if response == nil || len(response.DataSet) == 0 {
		return nil, false, nil
	}
	return &response.DataSet[0], true, nil
}

func testAccCheckDBInstanceExists(name string, target *udbapi.UDBInstanceSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("db instance id is empty")
		}

		client, err := testAccUDBClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccDBInstanceByID(client, item.Primary.ID)
		log.Printf("[INFO] db instance id %#v", item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("db instance %q is not found", item.Primary.ID)
		}
		*target = *instance
		return nil
	}
}

func testAccCheckDBInstanceAttributes(instance *udbapi.UDBInstanceSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if instance.DBId == "" {
			return fmt.Errorf("db instance id is empty")
		}
		return nil
	}
}

func testAccCheckDBInstanceDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_db_instance" {
			continue
		}

		client, err := testAccUDBClient()
		if err != nil {
			return err
		}
		instance, found, err := describeAccDBInstanceByID(client, item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if instance.DBId != "" {
			return fmt.Errorf("db instance still exist")
		}
	}
	return nil
}
