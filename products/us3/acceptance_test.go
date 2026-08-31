package us3_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	productus3 "github.com/terraform-providers/terraform-provider-ucloud/products/us3"
	providerucloud "github.com/terraform-providers/terraform-provider-ucloud/ucloud"
)

var (
	testAccProvider  *schema.Provider
	testAccProviders map[string]terraform.ResourceProvider
)

func init() {
	testAccProvider = providerucloud.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"ucloud": testAccProvider,
	}
}

func TestAccUCloudUS3Bucket_basic(t *testing.T) {
	var bucket ufile.UFileBucketSet
	bucketName := testAccBucketName("basic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "ucloud_us3_bucket.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckUS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUS3BucketConfig(bucketName, "private"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUS3BucketExists("ucloud_us3_bucket.foo", &bucket),
					testAccCheckUS3BucketAttributes(&bucket),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "name", bucketName),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "tag", "tf-acc"),
				),
			},
			{
				Config: testAccUS3BucketConfig(bucketName, "public"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUS3BucketExists("ucloud_us3_bucket.foo", &bucket),
					testAccCheckUS3BucketAttributes(&bucket),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "name", bucketName),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "type", "public"),
					resource.TestCheckResourceAttr("ucloud_us3_bucket.foo", "tag", "tf-acc"),
				),
			},
		},
	})
}

func TestAccUCloudUS3BucketsDataSource_basic(t *testing.T) {
	bucketName := testAccBucketName("data")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{{
			Config: testAccDataUS3BucketsConfig(bucketName),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckIDExists("data.ucloud_us3_buckets.foo"),
				resource.TestCheckResourceAttr("data.ucloud_us3_buckets.foo", "us3_buckets.#", "1"),
				resource.TestCheckResourceAttr("data.ucloud_us3_buckets.foo", "us3_buckets.0.name", bucketName),
				resource.TestCheckResourceAttr("data.ucloud_us3_buckets.foo", "us3_buckets.0.tag", "Default"),
			),
		}},
	})
}

func testAccPreCheck(t *testing.T) {
	for _, name := range []string{"UCLOUD_PUBLIC_KEY", "UCLOUD_PRIVATE_KEY", "UCLOUD_PROJECT_ID"} {
		if os.Getenv(name) == "" {
			t.Fatalf("%s must be set for acceptance tests", name)
		}
	}
	if os.Getenv("UCLOUD_REGION") == "" {
		log.Println("[INFO] Test: Using cn-bj2 as test region")
		if err := os.Setenv("UCLOUD_REGION", "cn-bj2"); err != nil {
			t.Fatalf("set UCLOUD_REGION: %v", err)
		}
	}
}

func testAccCheckIDExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find resource or data source: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("ID is not set")
		}
		return nil
	}
}

func testAccCheckUS3BucketExists(name string, target *ufile.UFileBucketSet) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		item, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if item.Primary.ID == "" {
			return fmt.Errorf("us3 bucket id is empty")
		}

		bucket, found, err := describeBucket(item.Primary.ID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("us3 bucket %q is not found", item.Primary.ID)
		}
		*target = *bucket
		return nil
	}
}

func testAccCheckUS3BucketAttributes(bucket *ufile.UFileBucketSet) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if bucket.BucketName == "" {
			return fmt.Errorf("us3 bucket id is empty")
		}
		return nil
	}
}

func testAccCheckUS3BucketDestroy(state *terraform.State) error {
	for _, item := range state.RootModule().Resources {
		if item.Type != "ucloud_us3_bucket" {
			continue
		}
		bucket, found, err := describeBucket(item.Primary.ID)
		if err != nil {
			return err
		}
		if found && bucket.BucketName != "" {
			return fmt.Errorf("us3 bucket still exists")
		}
	}
	return nil
}

func describeBucket(id string) (*ufile.UFileBucketSet, bool, error) {
	client, err := testUS3Client()
	if err != nil {
		return nil, false, err
	}
	req := client.NewDescribeBucketRequest()
	req.BucketName = ucloud.String(id)
	resp, err := client.DescribeBucket(req)
	if err != nil {
		if cloudErr, ok := err.(uerr.Error); ok && cloudErr.Code() == 15010 {
			return nil, false, nil
		}
		return nil, false, err
	}
	if resp == nil || len(resp.DataSet) == 0 {
		return nil, false, nil
	}
	return &resp.DataSet[0], true, nil
}

func testUS3Client() (*ufile.UFileClient, error) {
	runtime, ok := testAccProvider.Meta().(product.RuntimeV1)
	if !ok {
		return nil, fmt.Errorf("invalid provider runtime %T", testAccProvider.Meta())
	}
	client, err := runtime.ProductClient(productus3.Name, func(
		config *ucloud.Config,
		credential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		client := ufile.NewClient(config, credential)
		for _, handler := range handlers {
			client.AddHttpRequestHandler(handler)
		}
		return client
	})
	if err != nil {
		return nil, err
	}
	typed, ok := client.(*ufile.UFileClient)
	if !ok {
		return nil, fmt.Errorf("unexpected US3 client type %T", client)
	}
	return typed, nil
}

func testAccBucketName(purpose string) string {
	const bucketCharacters = "abcdefghijklmnopqrstuvwxyz0123456789"
	return fmt.Sprintf("tf-acc-us3-%s-%s", purpose, acctest.RandStringFromCharSet(10, bucketCharacters))
}

func testAccUS3BucketConfig(bucketName, bucketType string) string {
	return fmt.Sprintf(`
resource "ucloud_us3_bucket" "foo" {
  name = %q
  type = %q
  tag  = "tf-acc"
}
`, bucketName, bucketType)
}

func testAccDataUS3BucketsConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "ucloud_us3_bucket" "foo" {
  name = %q
  type = "private"
}

data "ucloud_us3_buckets" "foo" {
  name_regex = "${ucloud_us3_bucket.foo.name}"
}
`, bucketName)
}
