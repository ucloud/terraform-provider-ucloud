package ucloud

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UCloudClient is the ucloud openapi client
type UCloudClient struct {
	region    string
	projectId string

	config          *ucloud.Config
	credential      *auth.Credential
	requestHandlers []ucloud.HttpRequestHandler
	productClients  sync.Map
}

var _ product.RuntimeV1 = (*UCloudClient)(nil)

// ProductClient returns one lazily-created SDK client per product name.
// Products own the concrete SDK type and constructor; the provider runtime
// owns shared configuration, credentials, request handlers, and caching.
func (client *UCloudClient) ProductClient(
	name string,
	constructor product.ClientConstructor,
) (interface{}, error) {
	if name == "" {
		return nil, fmt.Errorf("product client name is empty")
	}
	if constructor == nil {
		return nil, fmt.Errorf("product client %q constructor is nil", name)
	}

	if productClient, ok := client.productClients.Load(name); ok {
		return productClient, nil
	}
	if client.config == nil || client.credential == nil {
		return nil, fmt.Errorf("product client %q requested before provider configuration", name)
	}

	config := *client.config
	credential := *client.credential
	handlers := append([]ucloud.HttpRequestHandler(nil), client.requestHandlers...)
	productClient := constructor(&config, &credential, handlers)
	if isNilProductClient(productClient) {
		return nil, fmt.Errorf("product client %q constructor returned nil", name)
	}
	actual, loaded := client.productClients.LoadOrStore(name, productClient)
	if loaded {
		return actual, nil
	}
	return productClient, nil
}

func isNilProductClient(value interface{}) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
