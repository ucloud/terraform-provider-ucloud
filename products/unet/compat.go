package unet

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
	defaultTag        = "Default"
	statusPending     = "pending"
	statusInitialized = "initialized"
	eipStatusFree     = "free"
	eipStatusUsed     = "used"

	resourceTypeInstance  = "instance"
	resourceTypeLb        = "lb"
	resourceTypeBareMetal = "baremetal"
	eipResourceTypeUHost  = "uhost"
	eipResourceTypeULB    = "ulb"
	eipResourceTypeUPHost = "upm"
)

var (
	validateDuration = validation.IntBetween(0, 9)
	validateName     = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
		"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	validateTag = validation.StringMatch(
		regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
		"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
	)
	upperCvt         = newUpperConverter(nil)
	upperCamelCvt    = newUpperCamelConverter(nil)
	lowerCaseProdCvt = newStringConverter(map[string]string{
		resourceTypeInstance:  eipResourceTypeUHost,
		resourceTypeLb:        eipResourceTypeULB,
		resourceTypeBareMetal: eipResourceTypeUPHost,
	})
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

func checkStringIn(value string, choices []string) error {
	for _, choice := range choices {
		if value == choice {
			return nil
		}
	}
	return fmt.Errorf("should be one of %q, got %q", strings.Join(choices, ","), value)
}

func isStringIn(value string, choices []string) bool {
	return checkStringIn(value, choices) == nil
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

	_ = os.Remove(absPath)
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
	// Keep legacy behavior: callers intentionally ignore write failures.
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

func validatePortRange(value interface{}, key string) (warnings []string, errors []error) {
	input := value.(string)
	parts := strings.Split(input, "-")
	if len(parts) > 2 {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", key, input))
	}

	fromPort, err := strconv.Atoi(parts[0])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", key, input))
	}
	if fromPort < 1 || fromPort > 65535 {
		errors = append(errors, fmt.Errorf("%q is invalid, should between 1-65535, got %q", key, input))
	}
	if len(parts) == 1 {
		return warnings, errors
	}

	toPort, err := strconv.Atoi(parts[1])
	if err != nil {
		errors = append(errors, fmt.Errorf("%q is invalid, should like a number or number1-number2, got %q", key, input))
	}
	if toPort < 1 || toPort > 65535 {
		errors = append(errors, fmt.Errorf("%q is invalid, should between 1-65535, got %q", key, input))
	}
	if toPort <= fromPort {
		errors = append(errors, fmt.Errorf("%q is invalid, for number1|number2, number2 must be greater than number1, got %q", key, input))
	}
	return warnings, errors
}

type associationInfo struct {
	PrimaryType  string
	PrimaryId    string
	ResourceType string
	ResourceId   string
}

var associationPattern = regexp.MustCompile(`^([^$]+)#([^:]+):([^$]+)#(.+)$`)

func parseAssociationInfo(id string) (*associationInfo, error) {
	matched := associationPattern.FindStringSubmatch(id)
	if len(matched) < 5 {
		return nil, fmt.Errorf("invalid identity of association")
	}
	return &associationInfo{
		PrimaryType: matched[1], PrimaryId: matched[2], ResourceType: matched[3], ResourceId: matched[4],
	}, nil
}
