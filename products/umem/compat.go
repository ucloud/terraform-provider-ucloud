package umem

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	notFoundCode = "Notfound"
	defaultTag   = "Default"

	statusPending     = "pending"
	statusInitialized = "initialized"
	statusRunning     = "Running"
)

type providerError struct {
	errorCode string
	message   string
}

func (err *providerError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func newNotFoundError(message string) error {
	return &providerError{errorCode: notFoundCode, message: message}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*providerError)
	return ok && (providerErr.errorCode == notFoundCode ||
		strings.Contains(strings.ToLower(providerErr.message), strings.ToLower(notFoundCode)))
}

type stringConverter struct {
	values  map[string]string
	reverse map[string]string
}

func newStringConverter(values map[string]string) stringConverter {
	reverse := make(map[string]string, len(values))
	for key, value := range values {
		reverse[value] = key
	}
	return stringConverter{values: values, reverse: reverse}
}

func (converter stringConverter) convert(value string) string {
	if converted, ok := converter.values[value]; ok {
		return converted
	}
	return value
}

func (converter stringConverter) unconvert(value string) string {
	if converted, ok := converter.reverse[value]; ok {
		return converted
	}
	return value
}

var upperCamelCvt = newStringConverter(map[string]string{
	"Year":    "year",
	"Month":   "month",
	"Dynamic": "dynamic",
})

var validateDuration = validation.IntBetween(0, 9)

var validateKVStoreInstanceName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9-_]{6,63}$`),
	"expected value to be have 1 - 63 characters and only support english, numbers, '-', '_'",
)

var (
	instancePasswordUpperPattern = regexp.MustCompile(`[A-Z]`)
	instancePasswordLowerPattern = regexp.MustCompile(`[a-z]`)
	instancePasswordNumPattern   = regexp.MustCompile(`[0-9]`)
	kvstorePasswordSpecial       = regexp.MustCompile(`[-_]`)
	kvstorePasswordPattern       = regexp.MustCompile(`^[A-Za-z0-9-_]{6,36}$`)
)

func validateKVStoreInstancePassword(value interface{}, key string) (warnings []string, errors []error) {
	password := value.(string)
	if !kvstorePasswordPattern.MatchString(password) {
		errors = append(errors, fmt.Errorf("%q is invalid, should have between 6-36 characters and any characters must be legal, got %q", key, password))
	}

	categoryCount := 0
	if instancePasswordUpperPattern.MatchString(password) {
		categoryCount++
	}
	if instancePasswordLowerPattern.MatchString(password) {
		categoryCount++
	}
	if instancePasswordNumPattern.MatchString(password) {
		categoryCount++
	}
	if kvstorePasswordSpecial.MatchString(password) {
		categoryCount++
	}
	if categoryCount < 3 {
		errors = append(errors, fmt.Errorf("%q is invalid, should have least 3 items of Capital letters, small letter, numbers and special characters, got %q", key, password))
	}
	return warnings, errors
}

func notEmptyStringInSet(value string) bool {
	return value != ""
}

func timestampToString(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

type redisInstanceType struct {
	Engine string
	Type   string
	Memory int
}

var availableRedisType = []string{"master", "distributed"}

func parseRedisInstanceType(value string) (*redisInstanceType, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("redis instance type is invalid, should like redis-xx-1, got %s", value)
	}
	if parts[0] != "redis" {
		return nil, fmt.Errorf("redis instance type is invalid, the engine of instance type must be %q", "redis")
	}
	if err := checkStringIn(parts[1], availableRedisType); err != nil {
		return nil, fmt.Errorf("redis instance type is invalid, the type of instance type  %s", err)
	}
	memory, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("redis instance type is invalid, the memory of instance type %s", err)
	}
	return &redisInstanceType{Engine: parts[0], Type: parts[1], Memory: memory}, nil
}

type memcacheInstanceType struct {
	Engine string
	Type   string
	Memory int
}

func parseMemcacheInstanceType(value string) (*memcacheInstanceType, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("memcache instance type is invalid, should like memcache-xx-1, got %s", value)
	}
	if parts[0] != "memcache" {
		return nil, fmt.Errorf("memcache instance type is invalid, the engine of instance type must be %q", "memcache")
	}
	if parts[1] != "master" {
		return nil, fmt.Errorf("memcache instance type is invalid, the type of instance type must be %q", "master")
	}
	memory, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("memcache instance type is invalid, the memory of instance type %s", err)
	}
	return &memcacheInstanceType{Engine: parts[0], Type: parts[1], Memory: memory}, nil
}

func isStringIn(value string, choices []string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func checkStringIn(value string, choices []string) error {
	if isStringIn(value, choices) {
		return nil
	}
	return fmt.Errorf("should be one of %q, got %q", strings.Join(choices, ","), value)
}

func validateRedisInstanceType(value interface{}, key string) (warnings []string, errors []error) {
	if _, err := parseRedisInstanceType(value.(string)); err != nil {
		errors = append(errors, err)
	}
	return warnings, errors
}

func validateMemcacheInstanceType(value interface{}, key string) (warnings []string, errors []error) {
	if _, err := parseMemcacheInstanceType(value.(string)); err != nil {
		errors = append(errors, err)
	}
	return warnings, errors
}

func validateAll(validators ...schema.SchemaValidateFunc) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		for _, validator := range validators {
			validatorWarnings, validatorErrors := validator(value, key)
			warnings = append(warnings, validatorWarnings...)
			errors = append(errors, validatorErrors...)
		}
		return warnings, errors
	}
}
