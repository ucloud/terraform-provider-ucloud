package ucloud

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ucloud": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UCLOUD_PUBLIC_KEY"); v == "" {
		t.Fatal("UCLOUD_PUBLIC_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("UCLOUD_PRIVATE_KEY"); v == "" {
		t.Fatal("UCLOUD_PRIVATE_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("UCLOUD_REGION"); v == "" {
		log.Println("[INFO] Test: Using cn-sh2 as test region")
		os.Setenv("UCLOUD_REGION", "cn-sh2")
	}
	if v := os.Getenv("UCLOUD_PROJECT_ID"); v == "" {
		t.Fatal("UCLOUD_PROJECT_ID must be set for acceptance tests")
	}
}

func testAccCheckIDExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find resource or data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not be set")
		}
		return nil
	}
}

func testAccCheckFileExists(filePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file %s is not exists", filePath)
		} else if err != nil {
			return err
		}

		return nil
	}
}
