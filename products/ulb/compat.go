package ulb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	defaultTag = "Default"

	statusPending     = "pending"
	statusInitialized = "initialized"

	resourceTypeInstance = "instance"
	lbResourceTypeUHost  = "UHost"
	eipResourceTypeULB   = "ulb"
	eipResourceTypeUHost = "uhost"
	lbMatchTypePath      = "Path"
	lbMatchTypeDomain    = "Domain"
)

var (
	validateName = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateTag = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
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

type upperCaseConverter struct{}

func (upperCaseConverter) convert(value string) string {
	return strings.ToLower(value)
}

func (upperCaseConverter) unconvert(value string) string {
	return strings.ToUpper(value)
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

type intConverter map[int]string

func (converter intConverter) convert(value int) string {
	if converted, ok := converter[value]; ok {
		return converted
	}
	return "unknown"
}

var (
	listenerStatusCvt = intConverter{
		0: "allNormal",
		1: "partNormal",
		2: "allException",
	}
	lbAttachmentStatusCvt = intConverter{
		0: "normalRunning",
		1: "exceptionRunning",
	}
	lbBackendCaseProdCvt = newStringConverter(map[string]string{
		"instance": "UHost",
	})
	upperCvt      = upperCaseConverter{}
	upperCamelCvt = newStringConverter(map[string]string{
		"Year":               "year",
		"Month":              "month",
		"Dynamic":            "dynamic",
		"RequestProxy":       "request_proxy",
		"PacketsTransmit":    "packets_transmit",
		"Roundrobin":         "roundrobin",
		"WeightRoundrobin":   "weight_roundrobin",
		"Source":             "source",
		"SourcePort":         "source_port",
		"ConsistentHash":     "consistent_hash",
		"ConsistentHashPort": "consistent_hash_port",
		"Leastconn":          "leastconn",
		"ServerInsert":       "server_insert",
		"UserDefined":        "user_defined",
		"None":               "none",
		"Port":               "port",
		"Path":               "path",
	})
)

func stateFuncTag(value interface{}) string {
	if len(value.(string)) == 0 {
		return defaultTag
	}
	return value.(string)
}

func schemaSetToStringSlice(value interface{}) []string {
	result := make([]string, 0)
	for _, item := range value.(*schema.Set).List() {
		result = append(result, item.(string))
	}
	return result
}

func isStringIn(value string, choices []string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func notEmptyStringInSet(value string) bool {
	return value != ""
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

	os.Remove(absPath)
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
	// Keep the legacy behavior: callers intentionally ignore write failures.
	_ = ioutil.WriteFile(absPath, content, 422)
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

func upperCamelConvert(value string) string {
	if converted, ok := upperCamelCvt.values[value]; ok {
		return converted
	}
	if value == "" {
		return ""
	}
	return lowerCamelToLower(strings.ToLower(value[:1]) + value[1:])
}

func upperCamelUnconvert(value string) string {
	if converted, ok := upperCamelCvt.reverse[value]; ok {
		return converted
	}
	if value == "" {
		return ""
	}
	output := lowerToLowerCamel(value)
	return strings.ToUpper(output[:1]) + output[1:]
}

func upperConvert(value string) string {
	return strings.ToLower(value)
}

func upperUnconvert(value string) string {
	return strings.ToUpper(value)
}

func lowerCamelToLower(input string) string {
	var state int
	var words []string
	buffer := strings.Builder{}
	for i := 0; i < len(input); i++ {
		current, next := input[i], lookAhead(input, i, 1)
		if next == 0 {
			buffer.WriteByte(toLowerASCII(current))
			words = append(words, buffer.String())
			buffer.Reset()
			break
		}
		if state == 0 {
			if isUpperASCII(next) {
				buffer.WriteByte(current)
				state = 1
				words = append(words, buffer.String())
				buffer.Reset()
			} else {
				buffer.WriteByte(current)
			}
			continue
		}
		if state == 1 {
			buffer.WriteByte(toLowerASCII(current))
			if isUpperASCII(next) {
				state = 3
			} else {
				state = 0
			}
			continue
		}
		if isUpperASCII(next) {
			buffer.WriteByte(toLowerASCII(current))
			continue
		}
		words = append(words, buffer.String())
		buffer.Reset()
		buffer.WriteByte(toLowerASCII(current))
		state = 0
	}
	return strings.Join(words, "_")
}

func lowerToLowerCamel(input string) string {
	parts := strings.Split(input, "_")
	var output strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		output.WriteString(strings.ToUpper(part[:1]))
		output.WriteString(part[1:])
	}
	result := output.String()
	if result == "" {
		return ""
	}
	return strings.ToLower(result[:1]) + result[1:]
}

func lookAhead(input string, index, forward int) byte {
	if len(input) <= index+forward {
		return 0
	}
	return input[index+forward]
}

func isUpperASCII(value byte) bool {
	return value >= 'A' && value <= 'Z'
}

func toLowerASCII(value byte) byte {
	if value >= 'A' && value <= 'Z' {
		return value + ('a' - 'A')
	}
	return value
}
