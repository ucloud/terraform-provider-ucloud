package ucloud

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*UCloudClient, error) {
	if os.Getenv("UCLOUD_PUBLIC_KEY") == "" {
		return nil, fmt.Errorf("empty UCLOUD_PUBLIC_KEY")
	}

	if os.Getenv("UCLOUD_PRIVATE_KEY") == "" {
		return nil, fmt.Errorf("empty UCLOUD_PRIVATE_KEY")
	}

	if os.Getenv("UCLOUD_PROJECT_ID") == "" {
		return nil, fmt.Errorf("empty UCLOUD_PROJECT_ID")
	}

	conf := &Config{
		Region:     region,
		PublicKey:  os.Getenv("UCLOUD_PUBLIC_KEY"),
		PrivateKey: os.Getenv("UCLOUD_PRIVATE_KEY"),
		ProjectId:  os.Getenv("UCLOUD_PROJECT_ID"),
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.Client()
	if err != nil {
		return nil, fmt.Errorf("error getting UCloud client")
	}

	return client, nil
}
