package ucloud

import (
	"sync"
	"testing"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestProductClientCachesByProductName(t *testing.T) {
	config := ucloud.NewConfig()
	config.Region = "cn-bj2"
	credential := auth.NewCredential()
	credential.PublicKey = "public-key"
	client := &UCloudClient{
		config:          &config,
		credential:      &credential,
		requestHandlers: []ucloud.HttpRequestHandler{nil},
	}

	created := 0
	constructor := func(
		productConfig *ucloud.Config,
		productCredential *auth.Credential,
		handlers []ucloud.HttpRequestHandler,
	) interface{} {
		created++
		if productConfig.Region != "cn-bj2" {
			t.Fatalf("constructor config region = %q", productConfig.Region)
		}
		if productCredential.PublicKey != "public-key" {
			t.Fatalf("constructor public key = %q", productCredential.PublicKey)
		}
		if len(handlers) != 1 {
			t.Fatalf("constructor handlers = %d, want 1", len(handlers))
		}
		productConfig.Region = "mutated"
		productCredential.PublicKey = "mutated"
		return &struct{ id int }{id: created}
	}

	first, err := client.ProductClient("example", product.ClientConstructor(constructor))
	if err != nil {
		t.Fatalf("first ProductClient() error = %v", err)
	}
	second, err := client.ProductClient("example", product.ClientConstructor(constructor))
	if err != nil {
		t.Fatalf("second ProductClient() error = %v", err)
	}
	if first != second {
		t.Fatal("ProductClient() did not return the cached client")
	}
	if created != 1 {
		t.Fatalf("constructor called %d times, want 1", created)
	}
	if config.Region != "cn-bj2" || credential.PublicKey != "public-key" {
		t.Fatalf("constructor mutated provider runtime: region=%q public_key=%q", config.Region, credential.PublicKey)
	}
}

func TestProductClientConcurrentAccessReturnsOneCachedClient(t *testing.T) {
	config := ucloud.NewConfig()
	credential := auth.NewCredential()
	client := &UCloudClient{config: &config, credential: &credential}
	constructor := func(*ucloud.Config, *auth.Credential, []ucloud.HttpRequestHandler) interface{} {
		return &struct{}{}
	}

	const callers = 32
	results := make(chan interface{}, callers)
	errors := make(chan error, callers)
	var wait sync.WaitGroup
	for index := 0; index < callers; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			value, err := client.ProductClient("example", constructor)
			if err != nil {
				errors <- err
				return
			}
			results <- value
		}()
	}
	wait.Wait()
	close(results)
	close(errors)

	for err := range errors {
		t.Errorf("ProductClient() error = %v", err)
	}
	var first interface{}
	for value := range results {
		if first == nil {
			first = value
			continue
		}
		if first != value {
			t.Fatal("concurrent callers received different cached clients")
		}
	}
}

func TestProductClientRejectsInvalidConstruction(t *testing.T) {
	tests := map[string]struct {
		client      *UCloudClient
		name        string
		constructor product.ClientConstructor
	}{
		"empty name": {
			client:      &UCloudClient{},
			constructor: func(*ucloud.Config, *auth.Credential, []ucloud.HttpRequestHandler) interface{} { return struct{}{} },
		},
		"nil constructor": {
			client: &UCloudClient{},
			name:   "example",
		},
		"unconfigured runtime": {
			client:      &UCloudClient{},
			name:        "example",
			constructor: func(*ucloud.Config, *auth.Credential, []ucloud.HttpRequestHandler) interface{} { return struct{}{} },
		},
		"nil client": {
			client: func() *UCloudClient {
				config := ucloud.NewConfig()
				credential := auth.NewCredential()
				return &UCloudClient{config: &config, credential: &credential}
			}(),
			name:        "example",
			constructor: func(*ucloud.Config, *auth.Credential, []ucloud.HttpRequestHandler) interface{} { return nil },
		},
		"typed nil client": {
			client: func() *UCloudClient {
				config := ucloud.NewConfig()
				credential := auth.NewCredential()
				return &UCloudClient{config: &config, credential: &credential}
			}(),
			name: "example",
			constructor: func(*ucloud.Config, *auth.Credential, []ucloud.HttpRequestHandler) interface{} {
				return (*struct{})(nil)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := tc.client.ProductClient(tc.name, tc.constructor); err == nil {
				t.Fatal("ProductClient() error = nil")
			}
		})
	}
}
