package product

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

var (
	productNamePattern        = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	terraformNamePattern      = regexp.MustCompile(`^ucloud_[a-z0-9_]+$`)
	terraformNamespacePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// Registration is the complete Terraform surface owned by one product.
// Products must use keyed fields so future optional fields remain source compatible.
type Registration struct {
	Name        string
	Resources   map[string]*schema.Resource
	DataSources map[string]*schema.Resource

	// Keep external product literals keyed so future optional fields remain
	// source compatible without requiring every product to change at once.
	reserved struct{}
}

// V1 is the stable product registration interface. Add new versions instead of
// adding methods here.
type V1 interface {
	Registration() Registration
}

// Binding fixes a product identity in core-owned provider wiring. Product
// packages cannot claim another Terraform namespace by changing Registration.Name.
type Binding struct {
	name                string
	adapter             V1
	terraformNamespaces []string
}

// BindOption changes core-owned registration constraints without expanding the
// product adapter interface.
type BindOption func(*Binding)

// Bind associates a core-owned product name with one product implementation.
func Bind(name string, adapter V1, options ...BindOption) Binding {
	binding := Binding{name: name, adapter: adapter}
	for _, option := range options {
		if option != nil {
			option(&binding)
		}
	}
	return binding
}

// WithTerraformNamespaces declares legacy Terraform name prefixes owned by a
// product. It is intended for core wiring when historical names do not match
// the UCloud product name, for example product "udisk" and "ucloud_disk_*".
func WithTerraformNamespaces(namespaces ...string) BindOption {
	values := append([]string(nil), namespaces...)
	return func(binding *Binding) {
		binding.terraformNamespaces = values
	}
}

// ClientConstructor lets a product own its SDK dependency while the provider
// runtime continues to own credentials, common configuration, and client caching.
// It must be side-effect free because concurrent first use can invoke it more
// than once before the runtime selects the cached client.
type ClientConstructor func(
	config *ucloud.Config,
	credential *auth.Credential,
	handlers []ucloud.HttpRequestHandler,
) interface{}

// RuntimeV1 is the stable runtime interface passed to product resources.
// Add a RuntimeV2 instead of adding methods here.
type RuntimeV1 interface {
	ProductClient(name string, constructor ClientConstructor) (interface{}, error)
}

// Register validates every product before atomically adding its Terraform
// registrations to the provider.
func Register(provider *schema.Provider, products ...Binding) error {
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}

	resources := cloneResources(provider.ResourcesMap)
	dataSources := cloneResources(provider.DataSourcesMap)
	names := make(map[string]struct{}, len(products))

	for _, binding := range products {
		if !productNamePattern.MatchString(binding.name) {
			return fmt.Errorf("bound product name %q must match %s", binding.name, productNamePattern)
		}
		if isNil(binding.adapter) {
			return fmt.Errorf("product %q adapter is nil", binding.name)
		}

		registration := binding.adapter.Registration()
		if registration.Name != binding.name {
			return fmt.Errorf("product declares name %q but is bound to %q by core", registration.Name, binding.name)
		}
		namespaces, err := bindingNamespaces(binding)
		if err != nil {
			return err
		}
		if err := validateRegistration(registration, namespaces); err != nil {
			return err
		}
		if _, exists := names[registration.Name]; exists {
			return fmt.Errorf("product %q is registered more than once", registration.Name)
		}
		names[registration.Name] = struct{}{}

		if err := addResources(resources, registration.Name, "resource", registration.Resources); err != nil {
			return err
		}
		if err := addResources(dataSources, registration.Name, "data source", registration.DataSources); err != nil {
			return err
		}
	}

	provider.ResourcesMap = resources
	provider.DataSourcesMap = dataSources
	return nil
}

// MustRegister turns invalid static product wiring into an immediate startup
// failure instead of exposing a partially registered provider.
func MustRegister(provider *schema.Provider, products ...Binding) {
	if err := Register(provider, products...); err != nil {
		panic(err)
	}
}

func validateRegistration(registration Registration, namespaces []string) error {
	if !productNamePattern.MatchString(registration.Name) {
		return fmt.Errorf("product name %q must match %s", registration.Name, productNamePattern)
	}
	if len(registration.Resources) == 0 && len(registration.DataSources) == 0 {
		return fmt.Errorf("product %q must register at least one resource or data source", registration.Name)
	}
	if err := validateResources(registration.Name, namespaces, "resource", registration.Resources); err != nil {
		return err
	}
	return validateResources(registration.Name, namespaces, "data source", registration.DataSources)
}

func validateResources(productName string, namespaces []string, kind string, resources map[string]*schema.Resource) error {
	for name, resource := range resources {
		if !terraformNamePattern.MatchString(name) {
			return fmt.Errorf("product %q %s name %q must start with ucloud_ and contain lowercase letters, numbers, or underscores", productName, kind, name)
		}
		if !matchesTerraformNamespace(name, namespaces) {
			return fmt.Errorf("product %q %s name %q must stay in Terraform namespaces %q", productName, kind, name, namespaces)
		}
		if resource == nil {
			return fmt.Errorf("product %q %s %q is nil", productName, kind, name)
		}
	}
	return nil
}

func bindingNamespaces(binding Binding) ([]string, error) {
	namespaces := binding.terraformNamespaces
	if len(namespaces) == 0 {
		namespaces = []string{strings.ReplaceAll(binding.name, "-", "_")}
	}
	result := make([]string, 0, len(namespaces))
	seen := make(map[string]struct{}, len(namespaces))
	for _, namespace := range namespaces {
		if !terraformNamespacePattern.MatchString(namespace) {
			return nil, fmt.Errorf("product %q Terraform namespace %q must match %s", binding.name, namespace, terraformNamespacePattern)
		}
		if _, exists := seen[namespace]; exists {
			return nil, fmt.Errorf("product %q contains duplicate Terraform namespace %q", binding.name, namespace)
		}
		seen[namespace] = struct{}{}
		result = append(result, "ucloud_"+namespace)
	}
	return result, nil
}

func matchesTerraformNamespace(name string, namespaces []string) bool {
	for _, namespace := range namespaces {
		if name == namespace || strings.HasPrefix(name, namespace+"_") {
			return true
		}
	}
	return false
}

func addResources(
	target map[string]*schema.Resource,
	productName string,
	kind string,
	resources map[string]*schema.Resource,
) error {
	for name, resource := range resources {
		if _, exists := target[name]; exists {
			return fmt.Errorf("product %q %s %q is already registered", productName, kind, name)
		}
		target[name] = resource
	}
	return nil
}

func cloneResources(resources map[string]*schema.Resource) map[string]*schema.Resource {
	cloned := make(map[string]*schema.Resource, len(resources))
	for name, resource := range resources {
		cloned[name] = resource
	}
	return cloned
}

func isNil(value interface{}) bool {
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
