package uk8s

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

	statusPending = "pending"
	statusRunning = "Running"
	statusRUNNING = "RUNNING"

	k8sClusterStatusCreateFailed = "CREATEFAILED"
	k8sClusterStatusDeleteFailed = "DELETEFAILED"
	k8sClusterStatusError        = "ERROR"
	k8sClusterStatusAbnormal     = "ABNORMAL"

	k8sNodeStatusError       = "Error"
	k8sNodeStatusReady       = "Ready"
	k8sNodeStatusInstallFail = "Install Fail"
	k8sNodeStatusStopped     = "Stopped"
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
		strings.Contains(strings.ToLower(providerErr.message), notFoundCode))
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

type upperCaseConverter struct{}

func (upperCaseConverter) convert(value string) string {
	return strings.ToLower(value)
}

func (upperCaseConverter) unconvert(value string) string {
	return strings.ToUpper(value)
}

var (
	upperCamelCvt = newStringConverter(map[string]string{
		"Year":    "year",
		"Month":   "month",
		"Dynamic": "dynamic",
	})
	upperCvt = upperCaseConverter{}

	validateDuration = validation.IntBetween(0, 9)
	validateName     = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateUImageName = validation.StringMatch(
		regexp.MustCompile(`^uimage-[A-Za-z0-9_]{5,50}$`),
		"expected value to start with uimage and followed by 5 - 50 characters and only support english, numbers, '_'",
	)
)

var (
	instancePasswordUpperPattern   = regexp.MustCompile(`[A-Z]`)
	instancePasswordLowerPattern   = regexp.MustCompile(`[a-z]`)
	instancePasswordNumPattern     = regexp.MustCompile(`[0-9]`)
	instancePasswordSpecialPattern = regexp.MustCompile("[`()~!@#$%^&*-+=_|{}\\[\\]:;'<>,.?/]")
	instancePasswordPattern        = regexp.MustCompile("^[A-Za-z0-9`()~!@#$%^&*-+=_|{}\\[\\]:;'<>,.?/]{8,30}$")
)

func validateInstanceType(value interface{}, key string) (warnings []string, errors []error) {
	if _, err := parseInstanceType(value.(string)); err != nil {
		errors = append(errors, err)
	}
	return warnings, errors
}

func validateInstancePassword(value interface{}, key string) (warnings []string, errors []error) {
	password := value.(string)
	if !instancePasswordPattern.MatchString(password) {
		errors = append(errors, fmt.Errorf("%q is invalid, should have between 8-30 characters and any characters must be legal, got %q", key, password))
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
	if instancePasswordSpecialPattern.MatchString(password) {
		categoryCount++
	}
	if categoryCount < 2 {
		errors = append(errors, fmt.Errorf("%q is invalid, should have least 2 items of capital letters, lower case letters, numbers and special characters, got %q", key, password))
	}
	return warnings, errors
}

func validateMod(num int) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		if value.(int)%num != 0 {
			errors = append(errors, fmt.Errorf("expected %q to be multiple of 10, got %d", key, value.(int)))
		}
		return warnings, errors
	}
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

func isStringIn(value string, choices []string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func timestampToString(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

type instanceType struct {
	CPU           int
	Memory        int
	HostType      string
	HostScaleType string
}

func parseInstanceType(value string) (*instanceType, error) {
	parts := strings.Split(value, "-")
	if len(parts) < 3 {
		return nil, fmt.Errorf("instance type is invalid, got %q", value)
	}
	if parts[1] == "customized" {
		return parseInstanceTypeByCustomize(parts...)
	}
	return parseInstanceTypeByNormal(parts...)
}

func parseInstanceTypeByCustomize(parts ...string) (*instanceType, error) {
	if len(parts) != 4 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-customized-1-2")
	}

	cpu, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}
	memory, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("memory count is invalid, please use a number")
	}

	if cpu != 1 && cpu%2 != 0 {
		return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
	}
	if memory != 1 && memory%2 != 0 {
		return nil, fmt.Errorf("expected the number of memory must be divisible by 2 without a remainder (except single memory), got %d", memory)
	}
	if cpu < 1 {
		return nil, fmt.Errorf("expected cpu to be more than 1, got %d", cpu)
	}
	if memory < 1 {
		return nil, fmt.Errorf("expected memory to be more than 1,got %d", memory)
	}
	if cpu/memory > 2 || memory/cpu > 12 || (cpu/memory == 2 && cpu%memory != 0) || (memory/cpu == 12 && memory%cpu != 0) {
		return nil, fmt.Errorf("the ratio of cpu to memory should be range of 2:1 ~ 1:12, got %d:%d", cpu, memory)
	}
	if (memory/cpu == 1 || memory/cpu == 2 || memory/cpu == 4 || memory/cpu == 8) && memory%cpu == 0 {
		return nil, fmt.Errorf("instance type is invalid, expected %q like %q,the Mode can be highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8", "n-Mode-CPU", "n-standard-1")
	}

	return &instanceType{
		CPU:           cpu,
		Memory:        memory * 1024,
		HostType:      parts[0],
		HostScaleType: parts[1],
	}, nil
}

var instanceTypeScaleMap = map[string]int{
	"highcpu":  1 * 1024,
	"basic":    2 * 1024,
	"standard": 4 * 1024,
	"highmem":  8 * 1024,
}

func parseInstanceTypeByNormal(parts ...string) (*instanceType, error) {
	if len(parts) != 3 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	}

	scale, ok := instanceTypeScaleMap[parts[1]]
	if !ok {
		return nil, fmt.Errorf("instance type is invalid, expected like %q,the Mode can be one of highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8, got %q ", "n-standard-1", parts[1])
	}
	cpu, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}
	if cpu != 1 && cpu%2 != 0 {
		return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
	}
	if cpu < 1 {
		return nil, fmt.Errorf("expected cpu to be more than 1, got %d", cpu)
	}

	return &instanceType{
		CPU:           cpu,
		Memory:        cpu * scale,
		HostType:      parts[0],
		HostScaleType: parts[1],
	}, nil
}
