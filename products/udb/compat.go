package udb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	defaultTag          = "Default"
	statusPending       = "pending"
	statusRunning       = "Running"
	statusShutdown      = "Shutdown"
	dbStatusShutoff     = "Shutoff"
	dbStatusRecoverFail = "RecoverFail"
	dbStatusFail        = "Fail"
	dbNVMeInstanceType  = "nvme"

	defaultPasswordNum = "012346789"
	defaultPasswordStr = "abcdefghijklmnopqrstuvwxyz"
	defaultPasswordSpe = "-_"
)

type providerError struct {
	errorCode string
	message   string
}

func (err *providerError) Error() string {
	return fmt.Sprintf("[ERROR] Terraform UCloud Provider Error: Code: %s Message: %s", err.errorCode, err.message)
}

func newNotFoundError(message string) error {
	return &providerError{errorCode: "Notfound", message: message}
}

func getNotFoundMessage(product, id string) string {
	return fmt.Sprintf("the specified %s %s is not found", product, id)
}

func isNotFoundError(err error) bool {
	providerErr, ok := err.(*providerError)
	return ok && (providerErr.errorCode == "Notfound" || strings.Contains(strings.ToLower(providerErr.message), "notfound"))
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

var (
	dbModeCvt = newStringConverter(map[string]string{
		"basic": "Normal",
		"ha":    "HA",
	})
	dbTypeCvt = newStringConverter(map[string]string{
		"nvme": "NVMe_SSD",
		"":     "SATA_SSD",
	})
	upperCamelCvt = newStringConverter(map[string]string{
		"Year":    "year",
		"Month":   "month",
		"Dynamic": "dynamic",
	})
	validateDuration = validation.IntBetween(0, 9)
	validateTag      = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateDBInstanceName = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.,\[\]:]{6,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.', ',', '[', ']', ':'",
	)
	validateDBInstanceBlackList = validation.StringMatch(
		regexp.MustCompile(`^[^.%]+\.([^.%]+|%)$`),
		fmt.Sprintf("expected element of %q should like %q or %q", "backup_black_list", "db.%", "dbname.tablename"),
	)
)

var instancePasswordPatterns = struct {
	upper   *regexp.Regexp
	lower   *regexp.Regexp
	number  *regexp.Regexp
	special *regexp.Regexp
	valid   *regexp.Regexp
}{
	upper:   regexp.MustCompile(`[A-Z]`),
	lower:   regexp.MustCompile(`[a-z]`),
	number:  regexp.MustCompile(`[0-9]`),
	special: regexp.MustCompile("[`()~!@#$%^&*-+=_|{}\\[\\]:;'<>,.?/]"),
	valid:   regexp.MustCompile("^[A-Za-z0-9`()~!@#$%^&*-+=_|{}\\[\\]:;'<>,.?/]{8,30}$"),
}

func validateDBInstanceType(value interface{}, key string) (warnings []string, errors []error) {
	if _, err := parseDBInstanceType(value.(string)); err != nil {
		errors = append(errors, err)
	}
	return warnings, errors
}

func validateDBInstancePassword(value interface{}, key string) (warnings []string, errors []error) {
	password := value.(string)
	if !instancePasswordPatterns.valid.MatchString(password) {
		errors = append(errors, fmt.Errorf("%q is invalid, should have between 8-30 characters and any characters must be legal, got %q", key, password))
	}

	categoryCount := 0
	if instancePasswordPatterns.upper.MatchString(password) {
		categoryCount++
	}
	if instancePasswordPatterns.lower.MatchString(password) {
		categoryCount++
	}
	if instancePasswordPatterns.number.MatchString(password) {
		categoryCount++
	}
	if instancePasswordPatterns.special.MatchString(password) {
		categoryCount++
	}
	if categoryCount < 3 {
		errors = append(errors, fmt.Errorf("%q is invalid, should have least 3 items of capital letters, lower case letters, numbers and special characters, got %q", key, password))
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

func stateFuncTag(value interface{}) string {
	if value.(string) == "" {
		return defaultTag
	}
	return value.(string)
}

func schemaListToStringSlice(value interface{}) []string {
	items := value.([]interface{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}

func schemaSetToStringSlice(value interface{}) []string {
	set := value.(*schema.Set)
	result := make([]string, 0, set.Len())
	for _, item := range set.List() {
		result = append(result, item.(string))
	}
	return result
}

func checkStringIn(value string, choices []string) error {
	for _, choice := range choices {
		if value == choice {
			return nil
		}
	}
	return fmt.Errorf("should be one of %q, got %q", strings.Join(choices, ","), value)
}

func isStringIn(value string, choices []string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func hashStringArray(values []string) string {
	var buffer bytes.Buffer
	for _, value := range values {
		buffer.WriteString(fmt.Sprintf("%s-", value))
	}
	return fmt.Sprintf("%d", hashcode.String(buffer.String()))
}

func timestampToString(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

func writeToFile(filePath string, data interface{}) error {
	absPath, err := getAbsPath(filePath)
	if err != nil {
		return err
	}

	var content []byte
	switch value := data.(type) {
	case string:
		content = []byte(value)
	default:
		content, err = json.MarshalIndent(value, "", "\t")
		if err != nil {
			return fmt.Errorf("MarshalIndent data %#v and got an error: %#v", data, err)
		}
	}
	os.Remove(absPath)
	ioutil.WriteFile(absPath, content, 422)
	return nil
}

func getAbsPath(filePath string) (string, error) {
	if strings.HasPrefix(filePath, "~") {
		currentUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("get current user got an error: %#v", err)
		}
		if currentUser.HomeDir != "" {
			filePath = strings.Replace(filePath, "~", currentUser.HomeDir, 1)
		}
	}
	return filepath.Abs(filePath)
}

type dbInstanceType struct {
	Engine string
	Mode   string
	Memory int
	Type   string
}

var availableDBEngine = []string{"mysql", "percona", "postgresql"}
var availableDBTypes = []string{"ha"}

func parseDBInstanceType(value string) (*dbInstanceType, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 3 && len(parts) != 4 {
		return nil, fmt.Errorf("db instance type is invalid, should like engine-mode-memory or engine-mode-type-memory, got %q", value)
	}
	if err := checkStringIn(parts[0], availableDBEngine); err != nil {
		return nil, fmt.Errorf("db instance type is invalid, the engine %s", err)
	}
	if err := checkStringIn(parts[1], availableDBTypes); err != nil {
		return nil, fmt.Errorf("db instance type is invalid, the type %s", err)
	}

	instanceType := ""
	memoryPart := parts[2]
	if len(parts) == 4 {
		instanceType = parts[2]
		if instanceType != dbNVMeInstanceType {
			return nil, fmt.Errorf("db instance type is invalid, the type of the machine architecture must be set %q, got %q", dbNVMeInstanceType, instanceType)
		}
		memoryPart = parts[3]
	}
	memory, err := strconv.Atoi(memoryPart)
	if err != nil {
		return nil, fmt.Errorf("db instance type is invalid, the memory %s", memoryPart)
	}
	return &dbInstanceType{Engine: parts[0], Mode: parts[1], Memory: memory, Type: instanceType}, nil
}
