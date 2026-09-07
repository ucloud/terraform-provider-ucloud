package providercompat

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const formatVersion = 1

// Contract contains the Terraform-visible behavior that must remain stable
// during product package migrations.
type Contract struct {
	FormatVersion int                         `json:"format_version"`
	Provider      map[string]SchemaContract   `json:"provider,omitempty"`
	Configure     string                      `json:"configure,omitempty"`
	Resources     map[string]ResourceContract `json:"resources,omitempty"`
	DataSources   map[string]ResourceContract `json:"data_sources,omitempty"`
}

type ResourceContract struct {
	Schema             map[string]SchemaContract `json:"schema,omitempty"`
	SchemaVersion      int                       `json:"schema_version,omitempty"`
	Create             string                    `json:"create,omitempty"`
	Read               string                    `json:"read,omitempty"`
	Update             string                    `json:"update,omitempty"`
	Delete             string                    `json:"delete,omitempty"`
	Exists             string                    `json:"exists,omitempty"`
	CustomizeDiff      string                    `json:"customize_diff,omitempty"`
	Importer           string                    `json:"importer,omitempty"`
	MigrateState       string                    `json:"migrate_state,omitempty"`
	StateUpgraders     []StateUpgraderContract   `json:"state_upgraders,omitempty"`
	Timeouts           map[string]string         `json:"timeouts,omitempty"`
	DeprecationMessage string                    `json:"deprecation_message,omitempty"`
}

type StateUpgraderContract struct {
	Version int    `json:"version"`
	Type    string `json:"type"`
	Upgrade string `json:"upgrade"`
}

type SchemaContract struct {
	Type          string           `json:"type"`
	ConfigMode    int              `json:"config_mode,omitempty"`
	Optional      bool             `json:"optional,omitempty"`
	Required      bool             `json:"required,omitempty"`
	Computed      bool             `json:"computed,omitempty"`
	ForceNew      bool             `json:"force_new,omitempty"`
	Sensitive     bool             `json:"sensitive,omitempty"`
	Description   string           `json:"description,omitempty"`
	MinItems      int              `json:"min_items,omitempty"`
	MaxItems      int              `json:"max_items,omitempty"`
	PromoteSingle bool             `json:"promote_single,omitempty"`
	Default       *string          `json:"default,omitempty"`
	DefaultFunc   string           `json:"default_func,omitempty"`
	InputDefault  string           `json:"input_default,omitempty"`
	DiffSuppress  string           `json:"diff_suppress,omitempty"`
	StateFunc     string           `json:"state_func,omitempty"`
	Set           string           `json:"set,omitempty"`
	Validate      string           `json:"validate,omitempty"`
	ComputedWhen  []string         `json:"computed_when,omitempty"`
	ConflictsWith []string         `json:"conflicts_with,omitempty"`
	ExactlyOneOf  []string         `json:"exactly_one_of,omitempty"`
	AtLeastOneOf  []string         `json:"at_least_one_of,omitempty"`
	Deprecated    string           `json:"deprecated,omitempty"`
	Removed       string           `json:"removed,omitempty"`
	Element       *ElementContract `json:"element,omitempty"`
}

type ElementContract struct {
	Kind     string                    `json:"kind"`
	Schema   *SchemaContract           `json:"schema,omitempty"`
	Resource map[string]SchemaContract `json:"resource,omitempty"`
}

// Marshal returns a deterministic, human-reviewable compatibility contract.
func Marshal(provider *schema.Provider) ([]byte, error) {
	contract, err := Build(provider)
	if err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal provider contract: %w", err)
	}
	return append(encoded, byte(10)), nil
}

// Build extracts the compatibility contract without evaluating dynamic
// defaults or making remote calls.
func Build(provider *schema.Provider) (Contract, error) {
	if provider == nil {
		return Contract{}, fmt.Errorf("provider is nil")
	}

	providerSchema, err := buildSchemaMap(provider.Schema)
	if err != nil {
		return Contract{}, fmt.Errorf("provider schema: %w", err)
	}
	resources, err := buildResourceMap(provider.ResourcesMap)
	if err != nil {
		return Contract{}, fmt.Errorf("resources: %w", err)
	}
	dataSources, err := buildResourceMap(provider.DataSourcesMap)
	if err != nil {
		return Contract{}, fmt.Errorf("data sources: %w", err)
	}

	return Contract{
		FormatVersion: formatVersion,
		Provider:      providerSchema,
		Configure:     functionName(provider.ConfigureFunc),
		Resources:     resources,
		DataSources:   dataSources,
	}, nil
}

func buildResourceMap(resources map[string]*schema.Resource) (map[string]ResourceContract, error) {
	result := make(map[string]ResourceContract, len(resources))
	for name, resource := range resources {
		if resource == nil {
			return nil, fmt.Errorf("%s is nil", name)
		}
		contract, err := buildResource(resource)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		result[name] = contract
	}
	return result, nil
}

func buildResource(resource *schema.Resource) (ResourceContract, error) {
	resourceSchema, err := buildSchemaMap(resource.Schema)
	if err != nil {
		return ResourceContract{}, err
	}

	stateUpgraders := make([]StateUpgraderContract, 0, len(resource.StateUpgraders))
	for _, upgrader := range resource.StateUpgraders {
		stateUpgraders = append(stateUpgraders, StateUpgraderContract{
			Version: upgrader.Version,
			Type:    upgrader.Type.GoString(),
			Upgrade: functionName(upgrader.Upgrade),
		})
	}

	var importer string
	if resource.Importer != nil {
		importer = functionName(resource.Importer.State)
	}

	return ResourceContract{
		Schema:             resourceSchema,
		SchemaVersion:      resource.SchemaVersion,
		Create:             functionName(resource.Create),
		Read:               functionName(resource.Read),
		Update:             functionName(resource.Update),
		Delete:             functionName(resource.Delete),
		Exists:             functionName(resource.Exists),
		CustomizeDiff:      functionName(resource.CustomizeDiff),
		Importer:           importer,
		MigrateState:       functionName(resource.MigrateState),
		StateUpgraders:     stateUpgraders,
		Timeouts:           buildTimeouts(resource.Timeouts),
		DeprecationMessage: resource.DeprecationMessage,
	}, nil
}

func buildSchemaMap(source map[string]*schema.Schema) (map[string]SchemaContract, error) {
	result := make(map[string]SchemaContract, len(source))
	for name, item := range source {
		if item == nil {
			return nil, fmt.Errorf("%s is nil", name)
		}
		contract, err := buildSchema(item)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		result[name] = contract
	}
	return result, nil
}

func buildSchema(source *schema.Schema) (SchemaContract, error) {
	element, err := buildElement(source.Elem)
	if err != nil {
		return SchemaContract{}, err
	}

	var defaultValue *string
	if source.Default != nil {
		encoded, err := json.Marshal(source.Default)
		if err != nil {
			return SchemaContract{}, fmt.Errorf("marshal default value of type %T: %w", source.Default, err)
		}
		value := string(encoded)
		defaultValue = &value
	}

	return SchemaContract{
		Type:          source.Type.String(),
		ConfigMode:    int(source.ConfigMode),
		Optional:      source.Optional,
		Required:      source.Required,
		Computed:      source.Computed,
		ForceNew:      source.ForceNew,
		Sensitive:     source.Sensitive,
		Description:   source.Description,
		MinItems:      source.MinItems,
		MaxItems:      source.MaxItems,
		PromoteSingle: source.PromoteSingle,
		Default:       defaultValue,
		DefaultFunc:   functionName(source.DefaultFunc),
		InputDefault:  source.InputDefault,
		DiffSuppress:  functionName(source.DiffSuppressFunc),
		StateFunc:     functionName(source.StateFunc),
		Set:           functionName(source.Set),
		Validate:      functionName(source.ValidateFunc),
		ComputedWhen:  sortedCopy(source.ComputedWhen),
		ConflictsWith: sortedCopy(source.ConflictsWith),
		ExactlyOneOf:  sortedCopy(source.ExactlyOneOf),
		AtLeastOneOf:  sortedCopy(source.AtLeastOneOf),
		Deprecated:    source.Deprecated,
		Removed:       source.Removed,
		Element:       element,
	}, nil
}

func buildElement(element interface{}) (*ElementContract, error) {
	switch typed := element.(type) {
	case nil:
		return nil, nil
	case *schema.Schema:
		if typed == nil {
			return nil, fmt.Errorf("schema element is nil")
		}
		contract, err := buildSchema(typed)
		if err != nil {
			return nil, err
		}
		return &ElementContract{Kind: "schema", Schema: &contract}, nil
	case *schema.Resource:
		if typed == nil {
			return nil, fmt.Errorf("resource element is nil")
		}
		resourceSchema, err := buildSchemaMap(typed.Schema)
		if err != nil {
			return nil, err
		}
		return &ElementContract{Kind: "resource", Resource: resourceSchema}, nil
	default:
		return nil, fmt.Errorf("unsupported element type %T", element)
	}
}

func buildTimeouts(timeouts *schema.ResourceTimeout) map[string]string {
	if timeouts == nil {
		return nil
	}
	result := make(map[string]string)
	addTimeout(result, "create", timeouts.Create)
	addTimeout(result, "read", timeouts.Read)
	addTimeout(result, "update", timeouts.Update)
	addTimeout(result, "delete", timeouts.Delete)
	addTimeout(result, "default", timeouts.Default)
	return result
}

func addTimeout(target map[string]string, name string, duration *time.Duration) {
	if duration != nil {
		target[name] = duration.String()
	}
}

func functionName(function interface{}) string {
	if function == nil {
		return ""
	}
	value := reflect.ValueOf(function)
	if value.Kind() != reflect.Func || value.IsNil() {
		return ""
	}
	resolved := runtime.FuncForPC(value.Pointer())
	if resolved == nil {
		return ""
	}
	return normalizeFunctionName(resolved.Name())
}

func normalizeFunctionName(name string) string {
	if slash := strings.LastIndex(name, "/"); slash >= 0 {
		name = name[slash+1:]
	}
	if dot := strings.Index(name, "."); dot >= 0 {
		name = name[dot+1:]
	}
	if strings.HasPrefix(name, "init.") {
		parts := strings.Split(name, ".")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	parts := strings.Split(name, ".")
	for index, part := range parts {
		if index > 0 && isGeneratedClosure(part) {
			return parts[index-1]
		}
	}
	return name
}

func isGeneratedClosure(name string) bool {
	for _, prefix := range []string{"func", "gowrap"} {
		if !strings.HasPrefix(name, prefix) || len(name) == len(prefix) {
			continue
		}
		for _, character := range name[len(prefix):] {
			if character < '0' || character > '9' {
				return false
			}
		}
		return true
	}
	return false
}

func sortedCopy(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
