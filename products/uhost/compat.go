package uhost

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
	defaultTag = "Default"

	statusPending = "pending"
	statusRunning = "Running"
	statusStopped = "Stopped"

	instanceStatusInitializing = "Initializing"
	instanceStatusStarting     = "Starting"
	instanceStatusRunning      = "Running"
	instanceStatusStopping     = "Stopping"
	instanceStatusStopped      = "Stopped"
	instanceStatusRebooting    = "Rebooting"
	instanceStatusInstallFail  = "Install Fail"
	instanceStatusResizeFail   = "ResizeFail"

	instanceBootDisksStatusNormal = "Normal"

	defaultPasswordNum = "012346789"
	defaultPasswordStr = "abcdefghijklmnopqrstuvwxyz"
	defaultPasswordSpe = "-_"

	eipResourceTypeUHost = "uhost"
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

type styleConverter interface {
	convertWithErr(string) (string, error)
	unconvertWithErr(string) (string, error)
	convert(string) string
	unconvert(string) string
}

type upperConverter struct{}

func newUpperConverter(_ map[string]string) styleConverter {
	return &upperConverter{}
}

func (cvt *upperConverter) convertWithErr(input string) (string, error) {
	if input != strings.ToUpper(input) {
		return "", fmt.Errorf("excepted input string is uppercase with underscore, got %q", input)
	}
	return cvt.convert(input), nil
}

func (cvt *upperConverter) convert(input string) string {
	return strings.ToLower(input)
}

func (cvt *upperConverter) unconvertWithErr(input string) (string, error) {
	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}
	return cvt.unconvert(input), nil
}

func (cvt *upperConverter) unconvert(input string) string {
	return strings.ToUpper(input)
}

type upperCamelConverter struct{}

func newUpperCamelConverter(_ map[string]string) styleConverter {
	return &upperCamelConverter{}
}

func (cvt *upperCamelConverter) convertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}
	if 'a' <= input[0] && input[0] <= 'z' {
		return "", fmt.Errorf("excepted upper camel should not be leading by lowercase character, got %q", input)
	}
	return lowerCamelToLower(strings.ToLower(input[:1]) + input[1:]), nil
}

func (cvt *upperCamelConverter) convert(input string) string {
	output, _ := cvt.convertWithErr(input)
	return output
}

func (cvt *upperCamelConverter) unconvertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}
	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}
	output := lowerToLowerCamel(input)
	return strings.ToUpper(output[:1]) + output[1:], nil
}

func (cvt *upperCamelConverter) unconvert(input string) string {
	output, _ := cvt.unconvertWithErr(input)
	return output
}

func lowerCamelToLower(input string) string {
	// Keep the legacy acronym splitting behavior, for values such as UDBInstance.
	var state int
	var words []string
	buf := strings.Builder{}
	for i := 0; i < len(input); i++ {
		c, next := input[i], lookAhead(&input, i, 1)
		if next == 0 {
			buf.Write(bytes.ToLower([]byte{c}))
			words = append(words, buf.String())
			buf.Reset()
			break
		}

		if state == 0 {
			if 'A' <= next && next <= 'Z' {
				buf.WriteByte(c)
				state = 1
				words = append(words, buf.String())
				buf.Reset()
			} else {
				buf.WriteByte(c)
			}
			continue
		}

		if state == 1 {
			buf.WriteByte(toLowerASCII(c))
			if 'A' <= next && next <= 'Z' {
				state = 3
			} else {
				state = 0
			}
			continue
		}

		if 'A' <= next && next <= 'Z' {
			buf.WriteByte(toLowerASCII(c))
		} else {
			words = append(words, buf.String())
			buf.Reset()
			buf.WriteByte(toLowerASCII(c))
			state = 0
		}
	}

	return strings.Join(words, "_")
}

func toLowerASCII(value byte) byte {
	if 'A' <= value && value <= 'Z' {
		return value + ('a' - 'A')
	}
	return value
}

func lowerToLowerCamel(input string) string {
	inputList := strings.Split(input, "_")
	outputList := make([]string, len(inputList))
	for i, value := range inputList {
		outputList[i] = strings.Title(value)
	}
	output := strings.Join(outputList, "")
	return strings.ToLower(output[:1]) + output[1:]
}

func lookAhead(input *string, index, forward int) byte {
	if len(*input) <= index+forward {
		return 0
	}
	return (*input)[index+forward]
}

type boolConverter struct {
	values  map[bool]string
	reverse map[string]bool
}

func newBoolConverter(values map[bool]string) boolConverter {
	reverse := make(map[string]bool, len(values))
	for key, value := range values {
		reverse[value] = key
	}
	return boolConverter{values: values, reverse: reverse}
}

func (converter boolConverter) convert(value bool) string {
	return converter.values[value]
}

func (converter boolConverter) unconvert(value string) bool {
	return converter.reverse[value]
}

var upperCvt = newUpperConverter(nil)

var upperCamelCvt = newUpperCamelConverter(nil)

var boolCamelCvt = newBoolConverter(map[bool]string{
	true:  "Yes",
	false: "No",
})

var boolValueCvt = newBoolConverter(map[bool]string{
	true:  "True",
	false: "False",
})

func schemaSetToStringSlice(value interface{}) []string {
	result := make([]string, 0)
	for _, item := range value.(*schema.Set).List() {
		result = append(result, item.(string))
	}
	return result
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

func notEmptyStringInSet(value string) bool {
	return value != ""
}

type instanceType struct {
	CPU           int
	Memory        int
	HostType      string
	HostScaleType string
}

var instanceTypeScaleMap = map[string]int{
	"highcpu":  1 * 1024,
	"basic":    2 * 1024,
	"standard": 4 * 1024,
	"highmem":  8 * 1024,
}

func parseInstanceType(value string) (*instanceType, error) {
	split := strings.Split(value, "-")
	if len(split) < 3 {
		return nil, fmt.Errorf("instance type is invalid, got %q", value)
	}
	if split[1] == "customized" {
		return parseInstanceTypeByCustomize(split...)
	}
	return parseInstanceTypeByNormal(split...)
}

func parseInstanceTypeByCustomize(split ...string) (*instanceType, error) {
	if len(split) != 4 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-customized-1-2")
	}
	hostType := split[0]
	hostScaleType := split[1]
	cpu, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}
	memory, err := strconv.Atoi(split[3])
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
	return &instanceType{CPU: cpu, Memory: memory * 1024, HostType: hostType, HostScaleType: hostScaleType}, nil
}

func parseInstanceTypeByNormal(split ...string) (*instanceType, error) {
	if len(split) != 3 {
		return nil, fmt.Errorf("instance type is invalid, expected like n-standard-1")
	}
	hostType := split[0]
	hostScaleType := split[1]
	scale, ok := instanceTypeScaleMap[hostScaleType]
	if !ok {
		return nil, fmt.Errorf("instance type is invalid, expected like %q,the Mode can be one of highcpu, basic, standard, highmem when the ratio of cpu to memory is 1:1, 1:2, 1:4, 1:8, got %q ", "n-standard-1", hostScaleType)
	}
	cpu, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, fmt.Errorf("cpu count is invalid, please use a number")
	}
	if cpu != 1 && cpu%2 != 0 {
		return nil, fmt.Errorf("expected the number of cores of cpu must be divisible by 2 without a remainder (except single core), got %d", cpu)
	}
	if cpu < 1 {
		return nil, fmt.Errorf("expected cpu to be more than 1, got %d", cpu)
	}
	return &instanceType{CPU: cpu, Memory: cpu * scale, HostType: hostType, HostScaleType: hostScaleType}, nil
}

var validateDuration = validation.IntBetween(0, 9)

var validateName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{1,63}$`),
	"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
)

var validateIsolationGroupName = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.,\[\]:]{1,63}$`),
	"expected value to be 1 - 63 characters and only support chinese, english, numbers, '-', '_', '.', ',', '[', ']', ':'",
)

var validateTag = validation.StringMatch(
	regexp.MustCompile(`^[A-Za-z0-9\p{Han}-_.]{0,63}$`),
	"expected value to be 0 - 63 characters and only support chinese, english, numbers, '-', '_', '.'",
)

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

func validateInstanceType(value interface{}, key string) (warnings []string, errors []error) {
	if _, err := parseInstanceType(value.(string)); err != nil {
		errors = append(errors, err)
	}
	return warnings, errors
}

func validateInstanceLoginMode(diff *schema.ResourceDiff, _ interface{}) error {
	loginMode := diff.Get("login_mode").(string)
	keyPairID := diff.Get("key_pair_id").(string)
	rootPassword, hasRootPassword := diff.GetOkExists("root_password")
	rootPasswordValue := ""
	if hasRootPassword {
		rootPasswordValue = rootPassword.(string)
	}
	hasConfiguredRootPassword := hasRootPassword && (rootPasswordValue != "" || diff.HasChange("root_password"))
	return validateInstanceLoginModeValues(loginMode, keyPairID, rootPasswordValue, hasConfiguredRootPassword)
}

func validateInstanceLoginModeValues(loginMode, keyPairID, rootPassword string, hasRootPassword bool) error {
	switch loginMode {
	case "":
		if keyPairID != "" {
			return fmt.Errorf("%q is required when %q is set to %q", "login_mode", "key_pair_id", keyPairID)
		}
	case "Password":
		if keyPairID != "" {
			return fmt.Errorf("%q cannot be set when %q is %q", "key_pair_id", "login_mode", "Password")
		}
	case "KeyPair":
		if keyPairID == "" {
			return fmt.Errorf("%q is required when %q is %q", "key_pair_id", "login_mode", "KeyPair")
		}
		if hasRootPassword {
			return fmt.Errorf("%q cannot be set when %q is %q", "root_password", "login_mode", "KeyPair")
		}
	default:
		return fmt.Errorf("%q must be one of %q, %q, or empty, got %q", "login_mode", "Password", "KeyPair", loginMode)
	}
	return nil
}

var instancePasswordUpperPattern = regexp.MustCompile(`[A-Z]`)
var instancePasswordLowerPattern = regexp.MustCompile(`[a-z]`)
var instancePasswordNumPattern = regexp.MustCompile(`[0-9]`)
var instancePasswordSpecialPattern = regexp.MustCompile(`[\x60()~!@#$%^&*-+=_|{}\[\]:;'<>,.?/]`)
var instancePasswordPattern = regexp.MustCompile(`^[A-Za-z0-9\x60()~!@#$%^&*-+=_|{}\[\]:;'<>,.?/]{8,30}$`)

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

func stateFuncTag(value interface{}) string {
	if value.(string) == "" {
		return defaultTag
	}
	return value.(string)
}
